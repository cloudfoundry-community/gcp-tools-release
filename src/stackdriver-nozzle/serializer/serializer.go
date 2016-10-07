package serializer

import (
	"fmt"

	"cloud.google.com/go/logging"
	"errors"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/cloudfoundry-community/firehose-to-syslog/utils"
	"github.com/cloudfoundry/lager"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/genproto/googleapis/api/metric"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"path"
)

type Serializer interface {
	GetLog(*events.Envelope) *logging.Entry
	GetMetrics(*events.Envelope) ([]*monitoringpb.CreateTimeSeriesRequest, error)
	IsLog(*events.Envelope) bool
}

type cachingClientSerializer struct {
	cachingClient caching.Caching
	logger        lager.Logger
}

func NewSerializer(cachingClient caching.Caching, logger lager.Logger) Serializer {
	if cachingClient == nil {
		logger.Fatal("nilCachingClient", errors.New("caching client cannot be nil"))
	}

	cachingClient.GetAllApp()

	return &cachingClientSerializer{cachingClient, logger}
}

func (s *cachingClientSerializer) GetLog(e *events.Envelope) *logging.Entry {
	return &logging.Entry{
		Payload: e,
		Labels:  s.buildLabels(e),
	}
}

func (s *cachingClientSerializer) buildTimeSeriesRequest(projectID string, name string, value float64, eventTime int64, labels map[string]string) *monitoringpb.CreateTimeSeriesRequest {
	projectName := fmt.Sprintf("projects/%s", projectID)
	metricType := path.Join("custom.googleapis.com", name)

	return &monitoringpb.CreateTimeSeriesRequest{
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
								Seconds: eventTime,
							},
							StartTime: &timestamp.Timestamp{
								Seconds: eventTime,
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
}

func (s *cachingClientSerializer) GetMetrics(envelope *events.Envelope) ([]*monitoringpb.CreateTimeSeriesRequest, error) {
	labels := s.buildLabels(envelope)
	eventTime := envelope.GetTimestamp()

	switch envelope.GetEventType() {
	case events.Envelope_ValueMetric:
		valueMetric := envelope.GetValueMetric()
		return []*monitoringpb.CreateTimeSeriesRequest{
			s.buildTimeSeriesRequest(
				"todo", valueMetric.GetName(), valueMetric.GetValue(), eventTime, labels,
			),
		}, nil
	case events.Envelope_ContainerMetric:
		containerMetric := envelope.GetContainerMetric()
		return []*monitoringpb.CreateTimeSeriesRequest{
			s.buildTimeSeriesRequest(
				"todo", "diskBytesQuota", float64(containerMetric.GetDiskBytesQuota()), eventTime, labels,
			),
			s.buildTimeSeriesRequest(
				"todo", "instanceIndex", float64(containerMetric.GetInstanceIndex()), eventTime, labels,
			),
			s.buildTimeSeriesRequest(
				"todo", "cpuPercentage", float64(containerMetric.GetCpuPercentage()), eventTime, labels,
			),
			s.buildTimeSeriesRequest(
				"todo", "diskBytes", float64(containerMetric.GetDiskBytes()), eventTime, labels,
			),
			s.buildTimeSeriesRequest(
				"todo", "memoryBytes", float64(containerMetric.GetMemoryBytes()), eventTime, labels,
			),
			s.buildTimeSeriesRequest(
				"todo", "memoryBytesQuota", float64(containerMetric.GetMemoryBytesQuota()), eventTime, labels,
			),
		}, nil
	case events.Envelope_CounterEvent:
		counterEvent := envelope.GetCounterEvent()
		return []*monitoringpb.CreateTimeSeriesRequest{
			s.buildTimeSeriesRequest(
				"todo", counterEvent.GetName(), float64(counterEvent.GetTotal()), eventTime, labels,
			),
		}, nil
	default:
		return nil, fmt.Errorf("unknown event type: %v", envelope.EventType)
	}

}

func (s *cachingClientSerializer) IsLog(envelope *events.Envelope) bool {
	switch *envelope.EventType {
	case events.Envelope_ValueMetric, events.Envelope_ContainerMetric, events.Envelope_CounterEvent:
		return false
	default:
		return true
	}
}

func getApplicationId(envelope *events.Envelope) string {
	if envelope.GetEventType() == events.Envelope_HttpStartStop {
		return utils.FormatUUID(envelope.GetHttpStartStop().GetApplicationId())
	} else if envelope.GetEventType() == events.Envelope_LogMessage {
		return envelope.GetLogMessage().GetAppId()
	} else if envelope.GetEventType() == events.Envelope_ContainerMetric {
		return envelope.GetContainerMetric().GetApplicationId()
	} else {
		return ""
	}
}

func (s *cachingClientSerializer) buildLabels(envelope *events.Envelope) map[string]string {
	labels := map[string]string{}

	if envelope.Origin != nil {
		labels["origin"] = envelope.GetOrigin()
	}

	if envelope.EventType != nil {
		labels["eventType"] = envelope.GetEventType().String()
	}

	if envelope.Deployment != nil {
		labels["deployment"] = envelope.GetDeployment()
	}

	if envelope.Job != nil {
		labels["job"] = envelope.GetJob()
	}

	if envelope.Index != nil {
		labels["index"] = envelope.GetIndex()
	}

	if envelope.Ip != nil {
		labels["ip"] = envelope.GetIp()
	}

	if appId := getApplicationId(envelope); appId != "" {
		labels["applicationId"] = appId
		s.buildAppMetadataLabels(appId, labels, envelope)
	}

	return labels
}

func (s *cachingClientSerializer) buildAppMetadataLabels(appId string, labels map[string]string, envelope *events.Envelope) {
	app := s.cachingClient.GetAppInfo(appId)

	if app.Name != "" {
		labels["appName"] = app.Name
	}

	if app.SpaceName != "" {
		labels["spaceName"] = app.SpaceName
	}

	if app.SpaceGuid != "" {
		labels["spaceGuid"] = app.SpaceGuid
	}

	if app.OrgName != "" {
		labels["orgName"] = app.OrgName
	}

	if app.OrgGuid != "" {
		labels["orgGuid"] = app.OrgGuid
	}
}
