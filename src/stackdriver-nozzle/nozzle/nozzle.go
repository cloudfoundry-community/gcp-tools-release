package nozzle

import (
	"strings"

	"github.com/cloudfoundry-community/gcp-tools-release/src/stackdriver-nozzle/heartbeat"
	"github.com/cloudfoundry/sonde-go/events"
)

type PostMetricError struct {
	Errors []error
}

func (e *PostMetricError) Error() string {
	errors := []string{}
	for _, err := range e.Errors {
		errors = append(errors, err.Error())
	}
	return strings.Join(errors, "\n")
}

type Nozzle struct {
	LogHandler    Sink
	MetricHandler Sink
	Heartbeater   heartbeat.Heartbeater
}

func (n *Nozzle) HandleEvent(envelope *events.Envelope) error {
	var handler Sink
	if isLog(envelope) {
		handler = n.LogHandler
	} else {
		handler = n.MetricHandler
	}

	n.Heartbeater.AddCounter()
	return handler.Receive(envelope)
}

func isLog(envelope *events.Envelope) bool {
	switch *envelope.EventType {
	case events.Envelope_ValueMetric, events.Envelope_ContainerMetric, events.Envelope_CounterEvent:
		return false
	default:
		return true
	}
}
