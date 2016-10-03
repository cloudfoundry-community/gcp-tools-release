package main

import (
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/dev"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/firehose"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/nozzle"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/stackdriver"
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/config"
)

func main() {
	//todo: pull in logging library...
	kingpin.Parse()

	client := firehose.NewClient(*config.ApiEndpoint, *config.Username, *config.Password, *config.SkipSSLValidation)

	var err error
	if *config.Debug{
		println("Sending firehose to standard out")
		err = client.StartListening(&dev.StdOut{})
	} else {
		println("Sending firehose to Stackdriver")
		sdClient := stackdriver.NewClient(*config.ProjectID, *config.BatchCount, *config.BatchDuration)
		n := nozzle.Nozzle{StackdriverClient: sdClient}

		err = client.StartListening(&n)
	}

	if err != nil {
		panic(err)
	}

}
