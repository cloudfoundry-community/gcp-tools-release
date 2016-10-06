package main

import (
	"os"

	"github.com/cloudfoundry/sonde-go/events"
	"gopkg.in/alecthomas/kingpin.v2"
	"stackdriver-nozzle/firehose"
	"github.com/cloudfoundry-community/go-cfclient"
)

func main() {
	kingpin.Parse()

	apiEndpoint := os.Getenv("API_ENDPOINT")
	username := os.Getenv("FIREHOSE_USERNAME")
	password := os.Getenv("FIREHOSE_PASSWORD")
	_, skipSSLValidation := os.LookupEnv("SKIP_SSL_VALIDATION")

	cfConfig := &cfclient.Config{
		ApiAddress:        apiEndpoint,
		Username:          username,
		Password:          password,
		SkipSslValidation: skipSSLValidation}
	
	cfClient := cfclient.NewClient(cfConfig)

	client := firehose.NewClient(cfConfig, cfClient, nil)

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
