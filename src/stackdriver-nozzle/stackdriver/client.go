package stackdriver

import (
	"time"

	"stackdriver-nozzle/heartbeat"

	"cloud.google.com/go/logging"
	"cloud.google.com/go/monitoring/apiv3"
	"github.com/cloudfoundry/lager"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type Client interface {
	PostLog(log *logging.Entry)
	PostMetric(req *monitoringpb.CreateTimeSeriesRequest) error
}

type client struct {
	ctx          context.Context
	sdLogger     *logging.Logger
	metricClient *monitoring.MetricClient
	projectID    string
	logger       lager.Logger
	heartbeater  heartbeat.Heartbeater
}

const (
	logId                = "cf_logs"
	DefaultBatchCount    = "10"
	DefaultBatchDuration = "1s"
)

// TODO error handling #131310523
func NewClient(projectID string, batchCount int, batchDuration time.Duration, logger lager.Logger, hearbeater heartbeat.Heartbeater) Client {
	ctx := context.Background()

	sdLogger := newLogger(ctx, projectID, batchCount, batchDuration, logger)

	metricClient, err := monitoring.NewMetricClient(ctx, option.WithScopes("https://www.googleapis.com/auth/monitoring.write"))
	if err != nil {
		logger.Fatal("metricClient", err)
	}

	return &client{
		ctx:          ctx,
		sdLogger:     sdLogger,
		metricClient: metricClient,
		projectID:    projectID,
		logger:       logger,
		heartbeater:  hearbeater,
	}
}

func newLogger(ctx context.Context, projectID string, batchCount int, batchDuration time.Duration, logger lager.Logger) *logging.Logger {
	loggingClient, err := logging.NewClient(ctx, projectID)
	if err != nil {
		logger.Fatal("stackdriverClient", err)
	}

	loggingClient.OnError = func(err error) {
		logger.Fatal("stackdriverClientOnError", err)
	}

	return loggingClient.Logger(logId,
		logging.EntryCountThreshold(batchCount),
		logging.DelayThreshold(batchDuration),
	)
}

func (s *client) PostLog(log *logging.Entry) {
	s.heartbeater.AddCounter()
	s.sdLogger.Log(*log)
}

func (s *client) PostMetric(req *monitoringpb.CreateTimeSeriesRequest) error {
	s.heartbeater.AddCounter()
	//projectName := fmt.Sprintf("projects/%s", s.projectID)

	return s.metricClient.CreateTimeSeries(s.ctx, req)
}
