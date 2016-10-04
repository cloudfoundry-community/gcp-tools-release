package serializer

import (
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/cloudfoundry-community/firehose-to-syslog/utils"
	"github.com/cloudfoundry/sonde-go/events"
)

type Metric interface {
	GetName() string
	GetValue() float64
	GetLabels() map[string]string
}

type Log interface {
	GetPayload() interface{}
	GetLabels() map[string]string
}

type Serializer interface {
	GetLog(*events.Envelope) Log
	GetMetrics(*events.Envelope) []Metric
}

type cachingClientSerializer struct {
	cachingClient caching.Caching
}

type metric struct {
	name   string
	value  float64
	labels map[string]string
}

func (m *metric) GetName() string {
	return m.name
}
func (m *metric) GetValue() float64 {
	return m.value
}
func (m *metric) GetLabels() map[string]string {
	return m.labels
}

type log struct {
	payload interface{}
	labels  map[string]string
}

func (l *log) GetPayload() interface{} {
	return l.payload
}
func (l *log) GetLabels() map[string]string {
	return l.labels
}

func NewSerializer(cachingClient caching.Caching) Serializer {
	return &cachingClientSerializer{cachingClient}
}

func (s *cachingClientSerializer) GetLog(e *events.Envelope) Log {
	return &log{payload: e, labels: s.buildLabels(e)}
}

func (s *cachingClientSerializer) GetMetrics(e *events.Envelope) []Metric {
	return []Metric{&metric{
		name:   e.GetValueMetric().GetName(),
		value:  e.GetValueMetric().GetValue(),
		labels: s.buildLabels(e)}}
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
		labels["event_type"] = envelope.GetEventType().String()
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
		labels["application_id"] = appId
	}

	return labels
}
