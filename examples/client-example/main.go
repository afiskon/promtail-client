package main

import (
	"fmt"
	"github.com/afiskon/promtail-client/promtail"
	"log"
	"os"
	"time"
)

func displayUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s proto|json source-name job-name\n", os.Args[0])
	os.Exit(1)
}

func displayInvalidName(arg string) {
	fmt.Fprintf(os.Stderr, "Invalid %s: allowed characters are a-zA-Z0-9_-\n", arg)
	os.Exit(1)
}

func nameIsValid(name string) bool {
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			(c == '-') || (c == '_')) {
			return false
		}
	}
	return true
}

func main() {
	if len(os.Args) < 4 {
		displayUsage()
	}

	format := os.Args[1]
	source_name := os.Args[2]
	job_name := os.Args[3]
	if format != "proto" && format != "json" {
		displayUsage()
	}

	if !nameIsValid(source_name) {
		displayInvalidName("source-name")
	}

	if !nameIsValid(job_name) {
		displayInvalidName("job-name")
	}

	labels := "{source=\"" + source_name + "\",job=\"" + job_name + "\"}"
	conf := promtail.ClientConfig{
		PushURL:            "http://localhost:3100/api/prom/push",
		Labels:             labels,
		BatchWait:          5 * time.Second,
		BatchEntriesNumber: 10000,
		SendLevel:          promtail.INFO,
		PrintLevel:         promtail.ERROR,
	}

	var (
		loki promtail.Client
		err  error
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
		tstamp := time.Now().String()
		loki.Debugf("source=%s time = %s, i = %d\n", source_name, tstamp, i)
		loki.Infof("source=%s, time = %s, i = %d\n", source_name, tstamp, i)
		loki.Warnf("source = %s, time = %s, i = %d\n", source_name, tstamp, i)
		loki.Errorf("source = %s, time = %s, i = %d\n", source_name, tstamp, i)
		time.Sleep(1 * time.Second)
	}

	loki.Shutdown()
}
