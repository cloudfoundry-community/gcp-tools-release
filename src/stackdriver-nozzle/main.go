package main

import (
	"fmt"
	"github.com/evandbrown/gcp-tools-boshrelease/src/stackdriver-nozzle/firehose"
	"github.com/evandbrown/gcp-tools-boshrelease/src/stackdriver-nozzle/stackdriver"
)

func main() {
	sdClient := stackdriver.NewClient("")
	sdClient.Post("hello world 4")

	config := &firehose.ClientConfig{
		User: "admin",
		Password: "",
		ApiEndpoint: "",
		SkipSSLValidation: true,
	}

	client := firehose.NewClient(config)
	client.StartListening(&StdOut{})
}

type StdOut struct {
}

func (so *StdOut) Connect() bool {
	return true
}

func (so *StdOut) ShipEvents(event map[string]interface{}, whatIsThis string) {
	//fmt.Printf("%s: %+v\n\n", whatIsThis, event)

	eventType := event["event_type"]
	fmt.Printf("%s: %+v\n\n", eventType, event)
}
