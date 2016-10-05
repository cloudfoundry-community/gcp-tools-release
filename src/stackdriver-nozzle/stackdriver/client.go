package stackdriver

import (
	"time"

	"fmt"

	"path"

	"cloud.google.com/go/logging"
	"cloud.google.com/go/monitoring/apiv3"
	"github.com/cloudfoundry/lager"
	"github.com/golang/protobuf/ptypes/timestamp"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/api/metric"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type Client interface {
	PostLog(payload interface{}, labels map[string]string)
	PostMetric(name string, value float64, labels map[string]string) error
}

type client struct {
	ctx          context.Context
	sdLogger     *logging.Logger
	metricClient *monitoring.MetricClient
	projectID    string
	logger       lager.Logger
}

const (
	logId                = "cf_logs"
	DefaultBatchCount    = "10"
	DefaultBatchDuration = "1s"
)

// TODO error handling #131310523
func NewClient(projectID string, batchCount int, batchDuration time.Duration, logger lager.Logger) Client {
	ctx := context.Background()

	sdLogger, err := newLogger(ctx, projectID, batchCount, batchDuration)
	if err != nil {
		panic(err)
	}

	metricClient, err := monitoring.NewMetricClient(ctx, option.WithScopes("https://www.googleapis.com/auth/monitoring.write"))
	if err != nil {
		panic(err)
	}

	return &client{
		ctx:          ctx,
		sdLogger:     sdLogger,
		metricClient: metricClient,
		projectID:    projectID,
		logger:       logger,
	}
}

func newLogger(ctx context.Context, projectID string, batchCount int, batchDuration time.Duration) (*logging.Logger, error) {
	loggingClient, err := logging.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	loggingClient.OnError = func(err error) {
		panic(err)
	}

	logger := loggingClient.Logger(logId,
		logging.EntryCountThreshold(batchCount),
		logging.DelayThreshold(batchDuration),
	)
	return logger, nil
}

func (s *client) PostLog(payload interface{}, labels map[string]string) {
	entry := logging.Entry{
		Payload: payload,
		Labels:  labels,
	}
	s.sdLogger.Log(entry)
}

func (s *client) PostMetric(name string, value float64, labels map[string]string) error {
	projectName := fmt.Sprintf("projects/%s", s.projectID)
	metricType := path.Join("custom.googleapis.com", name)

	req := &monitoringpb.CreateTimeSeriesRequest{
		Name: projectName,
		TimeSeries: []*monitoringpb.TimeSeries{
			{
				Metric: &google_api.Metric{
					Type:   metricType,
					Labels: labels,
				},
				Points: []*monitoringpb.Point{
					{
						Interval: &monitoringpb.TimeInterval{
							EndTime: &timestamp.Timestamp{
								Seconds: time.Now().Unix(),
							},
							StartTime: &timestamp.Timestamp{
								Seconds: time.Now().Unix(),
							},
						},
						Value: &monitoringpb.TypedValue{
							Value: &monitoringpb.TypedValue_DoubleValue{
								DoubleValue: value,
							},
						},
					},
				},
			},
		},
	}
	err := s.metricClient.CreateTimeSeries(s.ctx, req)
	if err != nil {
		fmt.Printf("Name: %v Value: %f Error: %v\n", name, value, err.Error())
	}
	return nil
}
