package main

import (
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"stackdriver-nozzle/filter"
	"stackdriver-nozzle/firehose"
	"stackdriver-nozzle/nozzle"
	"stackdriver-nozzle/serializer"
	"stackdriver-nozzle/stackdriver"

	"fmt"
	"strings"

	"github.com/cloudfoundry/lager"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
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
	eventsFilter = kingpin.Flag("events", "events to subscribe to from firehose (comma separated)").
			Default("LogMessage,Error").
			OverrideDefaultFromEnvar("FIREHOSE_EVENTS").
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
	boltDatabasePath = kingpin.Flag("boltdb-path", "Bolt Database path").
				Default("cached-app-metadata.db").
				OverrideDefaultFromEnvar("BOLTDB_PATH").
				String()
)

func main() {
	kingpin.Parse()

	logger := lager.NewLogger("my-app")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	input := firehose.NewClient(*apiEndpoint, *username, *password, *skipSSLValidation, logger)

	sdClient := stackdriver.NewClient(*projectID, *batchCount, *batchDuration, logger)
	cachingClient := caching.NewCachingBolt(input.EnsureCfClient(), *boltDatabasePath)
	// Initialize the caching client with the state of the world
	cachingClient.GetAllApp()
	output := nozzle.Nozzle{
		StackdriverClient: sdClient,
		Serializer:        serializer.NewSerializer(cachingClient),
		Logger:            logger,
	}

	filteredOutput, err := filter.New(&output, strings.Split(*eventsFilter, ","))
	if err != nil {
		if invalidEvent, ok := err.(*filter.InvalidEvent); ok {
			logger.Fatal("invalidEvent", invalidEvent)
		} else {
			panic(err)
		}
	}

	logger.Info(fmt.Sprintf("Listening to event(s): '%v'", *eventsFilter))

	err = input.StartListening(filteredOutput)

	if err != nil {
		panic(err)
	}
}
