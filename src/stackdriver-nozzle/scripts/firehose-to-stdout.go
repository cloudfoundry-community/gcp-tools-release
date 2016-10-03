package main

import (
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/firehose"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/config"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	kingpin.Parse()

	client := firehose.NewClient(*config.ApiEndpoint, *config.Username, *config.Password, *config.SkipSSLValidation)

	err := client.StartListening(&StdOut{})
	if err != nil {
		panic(err)
	}
}

type StdOut struct{}

func (so *StdOut) HandleEvent(envelope *events.Envelope) error {
	println(envelope.String())
	return nil
}
