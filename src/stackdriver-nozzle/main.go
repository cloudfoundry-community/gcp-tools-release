package main

import (
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/config"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/firehose"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/nozzle"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/stackdriver"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	kingpin.Parse()

	client := firehose.NewClient(*config.ApiEndpoint, *config.Username, *config.Password, *config.SkipSSLValidation)

	sdClient := stackdriver.NewClient(*config.ProjectID, *config.BatchCount, *config.BatchDuration)
	n := nozzle.Nozzle{StackdriverClient: sdClient}

	err := client.StartListening(&n)

	if err != nil {
		panic(err)
	}

}
