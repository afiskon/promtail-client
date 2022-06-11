package promtail

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
)

const LOG_ENTRIES_CHAN_SIZE = 5000

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO  LogLevel = iota
	WARN  LogLevel = iota
	ERROR LogLevel = iota
	// Maximum level, disables sending or printing
	DISABLE LogLevel = iota
)

type ClientConfig struct {
	// E.g. http://localhost:3100/api/prom/push
	PushURL string
	// E.g. "{job=\"somejob\"}"
	Labels             string
	BatchWait          time.Duration
	BatchEntriesNumber int
	// Logs are sent to Promtail if the entry level is >= SendLevel
	SendLevel LogLevel
	// Logs are printed to stdout if the entry level is >= PrintLevel
	PrintLevel LogLevel
}

type Client interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	DebugfWithLabel(label, format string, args ...interface{})
	InfofWithLabel(label, format string, args ...interface{})
	WarnfWithLabel(label, format string, args ...interface{})
	ErrorfWithLabel(label, format string, args ...interface{})
	LogfWithLabel(label, format string, args ...interface{})
	Shutdown()
}

// http.Client wrapper for adding new methods, particularly sendJsonReq
type httpClient struct {
	parent http.Client
}

// A bit more convenient method for sending requests to the HTTP server
func (client *httpClient) sendJsonReq(method, url string, ctype string, reqBody []byte) (resp *http.Response, resBody []byte, err error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", ctype)

	resp, err = client.parent.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	resBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return resp, resBody, nil
}
