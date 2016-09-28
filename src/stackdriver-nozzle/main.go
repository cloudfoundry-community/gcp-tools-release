package main

import (
	"github.com/evandbrown/gcp-tools-boshrelease/src/stackdriver-nozzle/firehose"
	"github.com/evandbrown/gcp-tools-boshrelease/src/stackdriver-nozzle/stackdriver"

	"github.com/evandbrown/gcp-tools-boshrelease/src/stackdriver-nozzle/dev"
	"github.com/evandbrown/gcp-tools-boshrelease/src/stackdriver-nozzle/nozzle"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	debug = kingpin.Flag("debug", "send events to stdout").
		OverrideDefaultFromEnvar("DEBUG").
		Default(false).
		Bool()
	apiEndpoint = kingpin.Flag("api-endpoint",
		"CF API endpoint (use https://api.bosh-lite.com for BOSH Lite)").
		OverrideDefaultFromEnvar("API_ENDPOINT").
		Required().
		String()
	user = kingpin.Flag("user", "username").
		Default("admin").
		OverrideDefaultFromEnvar("FIREHOSE_USER").
		String()
	password = kingpin.Flag("password", "password").
			Default("admin").
			OverrideDefaultFromEnvar("FIREHOSE_PASSWORD").
			String()
	skipSSLValidation = kingpin.Flag("skip-ssl-validation", "please don't").
				Default("false").
				OverrideDefaultFromEnvar("SKIP_SSL_VALIDATION").
				Bool()
	projectID = kingpin.Flag("project-id", "gcp project id").
			OverrideDefaultFromEnvar("PROJECT_ID").
			String() //maybe we can get this from gcp env...? research
)

func main() {
	//todo: pull in logging library...
	kingpin.Parse()

	config := &firehose.ClientConfig{
		User:              *user,
		Password:          *password,
		ApiEndpoint:       *apiEndpoint,
		SkipSSLValidation: *skipSSLValidation,
	}

	client := firehose.NewClient(config)

	if *debug {
		println("Sending firehose to standard out")
		client.StartListening(&dev.StdOut{})
	} else {
		println("Sending firehose to Stackdriver")
		sdClient := stackdriver.NewClient(*projectID)
		//sdClient.Post("hello world 5")
		n := nozzle.Nozzle{StackdriverClient: sdClient}

		client.StartListening(&n)
	}
}
