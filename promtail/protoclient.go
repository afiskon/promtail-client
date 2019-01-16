package promtail

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/golang/snappy"
	"go-sandbox/loki-example/logproto"
	"log"
	"sync"
	"time"
)

type protoLogEntry struct {
	entry *logproto.Entry
	level LogLevel
}

type clientProto struct {
	config    *ClientConfig
	quit      chan struct{}
	entries   chan protoLogEntry
	waitGroup sync.WaitGroup
	client    httpClient
}

func NewClientProto(conf ClientConfig) (Client, error) {
	client := clientProto{
		config:  &conf,
		quit:    make(chan struct{}),
		entries: make(chan protoLogEntry, LOG_ENTRIES_CHAN_SIZE),
		client:  httpClient{},
	}

	client.waitGroup.Add(1)
	go client.run()

	return &client, nil
}

func (c *clientProto) Debugf(format string, args ...interface{}) {
	c.log(format, DEBUG, "Debug: ", args)
}

func (c *clientProto) Infof(format string, args ...interface{}) {
	c.log(format, INFO, "Info: ", args)
}

func (c *clientProto) Warnf(format string, args ...interface{}) {
	c.log(format, WARN, "Warn: ", args)
}

func (c *clientProto) Errorf(format string, args ...interface{}) {
	c.log(format, ERROR, "Error: ", args)
}

func (c *clientProto) log(format string, level LogLevel, prefix string, args ...interface{}) {
	if (level >= c.config.SendLevel) || (level >= c.config.PrintLevel) {
		now := time.Now().UnixNano()
		c.entries <- protoLogEntry{
			entry: &logproto.Entry{
				Timestamp: &timestamp.Timestamp{
					Seconds: now / int64(time.Second),
					Nanos:   int32(now % int64(time.Second)),
				},
				Line: fmt.Sprintf(prefix+format, args...),
			},
			level: level,
		}
	}
}

func (c *clientProto) Shutdown() {
	close(c.quit)
	c.waitGroup.Wait()
}

func (c *clientProto) run() {
	var batch []*logproto.Entry
	batchSize := 0
	maxWait := time.NewTimer(c.config.BatchWait)

	defer func() {
		if batchSize > 0 {
			c.send(batch)
		}

		c.waitGroup.Done()
	}()

	for {
		select {
		case <-c.quit:
			return
		case entry := <-c.entries:
			if entry.level >= c.config.PrintLevel {
				log.Print(entry.entry.Line)
			}

			if entry.level >= c.config.SendLevel {
				batch = append(batch, entry.entry)
				batchSize++
				if batchSize >= c.config.BatchEntriesNumber {
					c.send(batch)
					batch = []*logproto.Entry{}
					batchSize = 0
					maxWait.Reset(c.config.BatchWait)
				}
			}
		case <-maxWait.C:
			if batchSize > 0 {
				c.send(batch)
				batch = []*logproto.Entry{}
				batchSize = 0
				maxWait.Reset(c.config.BatchWait)
			}
		}
	}
}

func (c *clientProto) send(entries []*logproto.Entry) {
	var streams []*logproto.Stream
	streams = append(streams, &logproto.Stream{
		Labels:  c.config.Labels,
		Entries: entries,
	})

	req := logproto.PushRequest{
		Streams: streams,
	}

	buf, err := proto.Marshal(&req)
	if err != nil {
		log.Printf("promtail.ClientProto: unable to marshal: %s\n", err)
		return
	}

	buf = snappy.Encode(nil, buf)

	resp, body, err := c.client.sendJsonReq("POST", c.config.PushURL, "application/x-protobuf", buf)
	if err != nil {
		log.Printf("promtail.ClientProto: unable to send an HTTP request: %s\n", err)
		return
	}

	if resp.StatusCode != 204 {
		log.Printf("promtail.ClientProto: Unexpected HTTP status code: %d, message: %s\n", resp.StatusCode, body)
		return
	}
}
