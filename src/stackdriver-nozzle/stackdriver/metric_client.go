package stackdriver

import (
	"context"

	"cloud.google.com/go/monitoring/apiv3"
	"google.golang.org/api/option"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type MetricClient interface {
	Post(*monitoringpb.CreateTimeSeriesRequest) error
	CreateMetricDescriptor(request *monitoringpb.CreateMetricDescriptorRequest) error
}

func NewMetricClient() (MetricClient, error) {
	ctx := context.Background()
	sdMetricClient, err := monitoring.NewMetricClient(ctx, option.WithScopes("https://www.googleapis.com/auth/monitoring.write"))
	if err != nil {
		return nil, err
	}

	return &metricClient{
		sdMetricClient: sdMetricClient,
		ctx:            ctx,
	}, nil
}

type metricClient struct {
	sdMetricClient *monitoring.MetricClient
	ctx            context.Context
}

func (m *metricClient) Post(request *monitoringpb.CreateTimeSeriesRequest) error {
	return m.sdMetricClient.CreateTimeSeries(m.ctx, request)
}

func (m *metricClient) CreateMetricDescriptor(request *monitoringpb.CreateMetricDescriptorRequest) error {
	_, err := m.sdMetricClient.CreateMetricDescriptor(m.ctx, request)
	return err
}
