package dev

import (
	"errors"
	"fmt"
	"path"
	"time"

	"cloud.google.com/go/monitoring/apiv3"
	"github.com/golang/protobuf/ptypes/timestamp"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/api/metric"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

func ClearMetricDescriptors() {
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

func SendTimeSeries() {
	ctx := context.Background()
	metricClient, _ := monitoring.NewMetricClient(ctx, option.WithScopes("https://www.googleapis.com/auth/monitoring.write"))

	t := time.NewTicker(500 * time.Millisecond)
	errCount := 0
	for _ = range t.C {
		println(fmt.Sprintf("tick: %v, %v", errCount, time.Now().Second()))
		err := postMetric(metricClient, ctx, "andres_test", float64(1337), map[string]string{})
		if err != nil {
			errCount += 1
			fmt.Printf("A wild error #%v appeared: %v\n", errCount, err)
		}
		if errCount > 10 {
			panic(errors.New("too many errors"))
		}
	}
}

func postMetric(m *monitoring.MetricClient, ctx context.Context, name string, value float64, labels map[string]string) error {
	projectName := "projects/evandbrown17"
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
	err := m.CreateTimeSeries(ctx, req)
	return err
}
