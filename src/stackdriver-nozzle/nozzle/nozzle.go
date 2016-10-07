package nozzle

import (
	"strings"

	"stackdriver-nozzle/serializer"
	"stackdriver-nozzle/stackdriver"

	"github.com/cloudfoundry/sonde-go/events"

	"fmt"
	"google.golang.org/genproto/googleapis/monitoring/v3"
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
}

func (n *Nozzle) HandleEvent(envelope *events.Envelope) error {
	if n.Serializer.IsLog(envelope) {
		log := n.Serializer.GetLog(envelope)
		n.StackdriverClient.PostLog(log)
		return nil
	} else {
		metrics, err := n.Serializer.GetMetrics(envelope)
		if err != nil {
			return err
		}
		return n.postMetrics(metrics)
	}
}

func (n *Nozzle) postMetrics(metrics []*google_monitoring_v3.CreateTimeSeriesRequest) error {
	errorsCh := make(chan error)

	for _, metric := range metrics {
		n.postMetric(errorsCh, metric)
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

func (n *Nozzle) postMetric(errorsCh chan error, request *google_monitoring_v3.CreateTimeSeriesRequest) {
	go func() {
		err := n.StackdriverClient.PostMetric(request)
		if err != nil {
			errorsCh <- fmt.Errorf("request: %+v, error: %v", request, err.Error())
		} else {
			errorsCh <- nil
		}
	}()
}
