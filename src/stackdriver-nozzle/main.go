package main

import (
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/firehose"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/nozzle"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/stackdriver"
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/serializer"
)

var (
	apiEndpoint = kingpin.Flag("api-endpoint",
		"CF API endpoint (use https://api.bosh-lite.com for BOSH Lite)").
		OverrideDefaultFromEnvar("API_ENDPOINT").
		Required().
		String()
	username = kingpin.Flag("username", "username").
			Default("admin").
			OverrideDefaultFromEnvar("FIREHOSE_USERNAME").
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

	batchCount = kingpin.Flag("batch-count", "maximum number of entries to buffer").
			Default(stackdriver.DefaultBatchCount).
			OverrideDefaultFromEnvar("BATCH_COUNT").
			Int()
	batchDuration = kingpin.Flag("batch-duration", "maximum amount of seconds to buffer").
			Default(stackdriver.DefaultBatchDuration).
			OverrideDefaultFromEnvar("BATCH_DURATION").
			Duration()
)

func main() {
	kingpin.Parse()

	client := firehose.NewClient(*apiEndpoint, *username, *password, *skipSSLValidation)

	sdClient := stackdriver.NewClient(*projectID, *batchCount, *batchDuration)
	n := nozzle.Nozzle{
		StackdriverClient: sdClient,
		Serializer: serializer.NewSerializer(caching.NewCachingEmpty()),
	}

	err := client.StartListening(&n)

	if err != nil {
		panic(err)
	}

}
