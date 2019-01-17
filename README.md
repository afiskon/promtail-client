# promtail-client

Promtail client library. Promtail is an agent for [Loki](https://github.com/grafana/loki) logging system.

*Warning!* Please note, that at time of writing Loki and Promtail are not
production ready, at least until [issue #168](https://github.com/grafana/loki/issues/168#issuecomment-455169837) is open.

This library supports both JSON and Protobuf APIs.

Usage example:

```
package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"github.com/afiskon/promtail-client/promtail"
)

func displayUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s proto|json\n", os.Args[0])
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		displayUsage()
	}

	format := os.Args[1]
	if format != "proto" && format != "json" {
		displayUsage()
	}

	conf := promtail.ClientConfig{
		PushURL:            "http://localhost:3100/api/prom/push",
		Labels:             "{job=\"bar\"}",
		BatchWait:          5 * time.Second,
		BatchEntriesNumber: 10000,
		SendLevel: 			promtail.INFO,
		PrintLevel: 		promtail.ERROR,
	}

	var (
		loki promtail.Client
		err error
	)

	if format == "proto" {
		loki, err = promtail.NewClientProto(conf)
	} else {
		loki, err = promtail.NewClientJson(conf)
	}

	if err != nil {
		log.Printf("promtail.NewClient: %s\n", err)
		os.Exit(1)
	}

	for i := 1; i < 5; i++ {
		loki.Debugf("The time is %s, i = %d\n", time.Now().String(), i)
		loki.Infof("The time is %s, i = %d\n", time.Now().String(), i)
		loki.Warnf("The time is %s, i = %d\n", time.Now().String(), i)
		loki.Errorf("The time is %s, i = %d\n", time.Now().String(), i)
		time.Sleep(1 * time.Second)
	}

	loki.Shutdown()
}
```
