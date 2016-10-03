package config

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/stackdriver"
)

var (
	Debug = kingpin.Flag("debug", "send events to stdout").
		Default("false").
		OverrideDefaultFromEnvar("DEBUG").
		Bool()
	ApiEndpoint = kingpin.Flag("api-endpoint",
		"CF API endpoint (use https://api.bosh-lite.com for BOSH Lite)").
		OverrideDefaultFromEnvar("API_ENDPOINT").
		Required().
		String()
	Username = kingpin.Flag("username", "username").
		Default("admin").
		OverrideDefaultFromEnvar("FIREHOSE_USERNAME").
		String()
	Password = kingpin.Flag("password", "password").
		Default("admin").
		OverrideDefaultFromEnvar("FIREHOSE_PASSWORD").
		String()
	SkipSSLValidation = kingpin.Flag("skip-ssl-validation", "please don't").
		Default("false").
		OverrideDefaultFromEnvar("SKIP_SSL_VALIDATION").
		Bool()
	ProjectID = kingpin.Flag("project-id", "gcp project id").
		OverrideDefaultFromEnvar("PROJECT_ID").
		String() //maybe we can get this from gcp env...? research

	BatchCount = kingpin.Flag("batch-count", "maximum number of entries to buffer").
		Default(stackdriver.DefaultBatchCount).
		OverrideDefaultFromEnvar("BATCH_COUNT").
		Int()
	BatchDuration = kingpin.Flag("batch-duration", "maximum amount of seconds to buffer").
		Default(stackdriver.DefaultBatchDuration).
		OverrideDefaultFromEnvar("BATCH_DURATION").
		Duration()
)