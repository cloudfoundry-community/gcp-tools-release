package nozzle

import (
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/stackdriver"
)

type Nozzle struct {
	StackdriverClient stackdriver.Client
}

func (n *Nozzle) HandleEvent(envelope *events.Envelope) {
	labels := map[string]string{}
	switch envelope.GetEventType() {
	case events.Envelope_ValueMetric:
		valueMetric := envelope.GetValueMetric()
		name := valueMetric.GetName()
		value := valueMetric.GetValue()
		n.StackdriverClient.PostMetric(name, value)
	default:
		labels["event_type"] = envelope.GetEventType().String()
		n.StackdriverClient.PostLog(envelope, labels)
	}
}
