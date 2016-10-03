package nozzle

import (
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/stackdriver"
	"fmt"
	"strings"
)

type PostContainerMetricError struct {
	Errors map[string]error
}

func (e *PostContainerMetricError) Error() string {
	errors := []string{}
	for name, err := range(e.Errors) {
		errors = append(errors, fmt.Sprintf("%v: %v", name, err.Error()))
	}
	return strings.Join(errors, "\n")
}

type Nozzle struct {
	StackdriverClient stackdriver.Client
}

func (n *Nozzle) HandleEvent(eventsEnvelope *events.Envelope) error {
	envelope := Envelope{eventsEnvelope}
	labels := envelope.Labels()

	switch envelope.GetEventType() {
	case events.Envelope_ContainerMetric:
		return n.postContainerMetrics(envelope)
	case events.Envelope_ValueMetric:
		valueMetric := envelope.GetValueMetric()
		name := valueMetric.GetName()
		value := valueMetric.GetValue()

		err := n.StackdriverClient.PostMetric(name, value, labels)
		return err
	default:
		n.StackdriverClient.PostLog(envelope, labels)
		return nil
	}
}

func (n *Nozzle) postContainerMetrics(envelope Envelope) *PostContainerMetricError {
	containerMetric := envelope.GetContainerMetric()

	labels := envelope.Labels()
	labels["applicationId"] = containerMetric.GetApplicationId()

	errors := map[string]error{}

	err := n.StackdriverClient.PostMetric("diskBytesQuota", float64(containerMetric.GetDiskBytesQuota()), labels)
	if err != nil {
		errors["diskBytesQuota"] = err
	}

	err = n.StackdriverClient.PostMetric("instanceIndex", float64(containerMetric.GetInstanceIndex()), labels)
	if err != nil {
		errors["instanceIndex"] = err
	}

	err = n.StackdriverClient.PostMetric("cpuPercentage", float64(containerMetric.GetCpuPercentage()), labels)
	if err != nil {
		errors["cpuPercentage"] = err
	}

	err = n.StackdriverClient.PostMetric("diskBytes", float64(containerMetric.GetDiskBytes()), labels)
	if err != nil {
		errors["diskBytes"] = err
	}

	err = n.StackdriverClient.PostMetric("memoryBytes", float64(containerMetric.GetMemoryBytes()), labels)
	if err != nil {
		errors["memoryBytes"] = err
	}

	err = n.StackdriverClient.PostMetric("memoryBytesQuota", float64(containerMetric.GetMemoryBytesQuota()), labels)
	if err != nil {
		errors["memoryBytesQuota"] = err
	}

	if len(errors) == 0 {
		return nil
	} else {
		return &PostContainerMetricError{
			Errors: errors,
		}
	}
}