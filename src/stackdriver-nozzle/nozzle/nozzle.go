package nozzle

import (
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/stackdriver"
)

type Nozzle struct {
	StackdriverClient stackdriver.Client
}

func (n *Nozzle) HandleEvent(envelope *events.Envelope) {
	labels := map[string]string{
		"event_type": envelope.GetEventType().String(),
	}
	switch envelope.GetEventType() {
	case events.Envelope_ValueMetric:
		valueMetric := envelope.GetValueMetric()
		name := valueMetric.GetName()
		value := valueMetric.GetValue()

		err := n.StackdriverClient.PostMetric(name, value, labels)
		if err != nil {
			panic(err)
		}
	default:
		n.StackdriverClient.PostLog(envelope, labels)
	}
}
