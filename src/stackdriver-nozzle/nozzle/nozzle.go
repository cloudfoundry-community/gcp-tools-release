package nozzle

import (
	"strings"

	"github.com/cloudfoundry/sonde-go/events"
	"stackdriver-nozzle/serializer"
	"stackdriver-nozzle/stackdriver"

	"fmt"
	"github.com/cloudfoundry/lager"
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
	StackdriverClient stackdriver.Client
	Serializer        serializer.Serializer
	Logger            lager.Logger
}

func (n *Nozzle) HandleEvent(envelope *events.Envelope) error {
	if n.Serializer.IsLog(envelope) {
		log := n.Serializer.GetLog(envelope)
		n.StackdriverClient.PostLog(log.Payload, log.Labels)
		return nil
	} else {
		metrics := n.Serializer.GetMetrics(envelope)
		err := n.postMetrics(metrics)
		if err != nil {
			n.Logger.Error("metricError", err)
		}
		//TODO: should we return the error or only log it?
		return nil
	}
}

func (n *Nozzle) postMetrics(metrics []*serializer.Metric) error {
	errorsCh := make(chan error)

	for _, metric := range metrics {
		n.postMetric(errorsCh, metric.Name, metric.Value, metric.Labels)
	}

	errors := []error{}
	for range metrics {
		err := <-errorsCh
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) == 0 {
		return nil
	} else {
		return &PostMetricError{
			Errors: errors,
		}
	}
}

func (n *Nozzle) postMetric(errorsCh chan error, name string, value float64, labels map[string]string) {
	go func() {
		err := n.StackdriverClient.PostMetric(name, value, labels)
		if err != nil {
			errorsCh <- fmt.Errorf("Name %v: Value: %f, Error: %v", name, value, err.Error())
		} else {
			errorsCh <- nil
		}
	}()
}
