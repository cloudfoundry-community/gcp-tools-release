package mocks

import (
	"cloud.google.com/go/logging"
	"github.com/cloudfoundry/sonde-go/events"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type MockSerializer struct {
	GetLogFn     func(*events.Envelope) *logging.Entry
	GetMetricsFn func(*events.Envelope) ([]*monitoringpb.CreateTimeSeriesRequest, error)
	IsLogFn      func(*events.Envelope) bool
}

func (m *MockSerializer) GetLog(envelope *events.Envelope) *logging.Entry {
	if m.GetLogFn != nil {
		return m.GetLogFn(envelope)
	}
	return nil
}

func (m *MockSerializer) GetMetrics(envelope *events.Envelope) ([]*monitoringpb.CreateTimeSeriesRequest, error) {
	if m.GetMetricsFn != nil {
		return m.GetMetricsFn(envelope)
	}
	return nil, nil
}

func (m *MockSerializer) IsLog(envelope *events.Envelope) bool {
	if m.IsLogFn != nil {
		return m.IsLogFn(envelope)
	}
	return true
}
