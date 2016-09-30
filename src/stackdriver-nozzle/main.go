package main

import (
	"cloud.google.com/go/monitoring/apiv3"
	"github.com/evandbrown/gcp-tools-release/src/firehose-to-fluentd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/dev"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/firehose"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/nozzle"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/stackdriver"
	"google.golang.org/api/option"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	debug = kingpin.Flag("debug", "send events to stdout").
		Default("false").
		OverrideDefaultFromEnvar("DEBUG").
		Bool()
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
	//todo: pull in logging library...
	kingpin.Parse()

	client := firehose.NewClient(*apiEndpoint, *username, *password, *skipSSLValidation)

	if *debug {
		println("Sending firehose to standard out")
		err := client.StartListening(&dev.StdOut{})
		if err != nil {
			panic(err)
		}
	} else {
		println("Sending firehose to Stackdriver")
		sdClient := stackdriver.NewClient(*projectID, *batchCount, *batchDuration)
		n := nozzle.Nozzle{StackdriverClient: sdClient}

		client.StartListening(&n)
	}

}

func clearMetricDescriptors() {
	ctx := context.Background()
	metricClient, err := monitoring.NewMetricClient(ctx, option.WithScopes("https://www.googleapis.com/auth/monitoring"))
	if err != nil {
		panic(err)
	}

	req := &monitoringpb.ListMetricDescriptorsRequest{
		Name:   "projects/evandbrown17",
		Filter: "metric.type = starts_with(\"custom.googleapis.com/\")",
	}
	it := metricClient.ListMetricDescriptors(ctx, req)
	for {
		resp, err := it.Next()
		if err == monitoring.Done {
			break
		}
		if err != nil {
			// TODO: Handle error.
			panic(err)
		}

		req := &monitoringpb.DeleteMetricDescriptorRequest{
			Name: resp.Name,
		}
		err = metricClient.DeleteMetricDescriptor(ctx, req)
		if err != nil {
			panic(err)
		}
	}
}
