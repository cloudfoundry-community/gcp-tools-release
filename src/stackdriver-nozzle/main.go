package main

import (
	"fmt"
	"github.com/evandbrown/gcp-tools-boshrelease/src/stackdriver-nozzle/firehose"
	"github.com/evandbrown/gcp-tools-boshrelease/src/stackdriver-nozzle/stackdriver"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	apiEndpoint       = kingpin.Flag("api-endpoint", "Api endpoint address. For bosh-lite installation of CF: https://api.10.244.0.34.xip.io").
		OverrideDefaultFromEnvar("API_ENDPOINT").Required().String()
	user              = kingpin.Flag("user", "Admin user.").Default("admin").
		OverrideDefaultFromEnvar("FIREHOSE_USER").String()
	password          = kingpin.Flag("password", "Admin password.").Default("admin").
		OverrideDefaultFromEnvar("FIREHOSE_PASSWORD").String()
	skipSSLValidation = kingpin.Flag("skip-ssl-validation", "Please don't").Default("false").
		OverrideDefaultFromEnvar("SKIP_SSL_VALIDATION").Bool()
	projectID = kingpin.Flag("project-id", "gcp project id").
		OverrideDefaultFromEnvar("PROJECT_ID").String() //maybe we can get this from gcp env...? research
)

func main() {
	kingpin.Parse()

	sdClient := stackdriver.NewClient(*projectID)
	sdClient.Post("hello world 5")

	config := &firehose.ClientConfig{
		User:                 *user,
		Password:             *password,
		ApiEndpoint:          *apiEndpoint,
		SkipSSLValidation:    *skipSSLValidation,
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
