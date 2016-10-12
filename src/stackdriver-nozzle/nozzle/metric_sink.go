package nozzle

import (
	"fmt"
	"time"

	"github.com/cloudfoundry-community/gcp-tools-release/src/stackdriver-nozzle/stackdriver"
	"github.com/cloudfoundry/sonde-go/events"
)

func NewMetricSink(labelMaker LabelMaker, metricBuffer stackdriver.MetricsBuffer, unitParser UnitParser) Sink {
	return &metricSink{
		labelMaker:   labelMaker,
		metricBuffer: metricBuffer,
		unitParser:   unitParser,
	}
}

type metricSink struct {
	labelMaker   LabelMaker
	metricBuffer stackdriver.MetricsBuffer
	unitParser   UnitParser
}

func (ms *metricSink) Receive(envelope *events.Envelope) error {
	labels := ms.labelMaker.Build(envelope)

	timestamp := time.Duration(envelope.GetTimestamp())
	eventTime := time.Unix(
		int64(timestamp/time.Second),
		int64(timestamp%time.Second),
	)
	points := func(value float64) map[time.Time]float64 {
		return map[time.Time]float64{eventTime: value}
	}

	var metrics []stackdriver.Metric
	switch envelope.GetEventType() {
	case events.Envelope_ValueMetric:
		valueMetric := envelope.GetValueMetric()
		metrics = []stackdriver.Metric{{
			Name:   valueMetric.GetName(),
			Points: points(valueMetric.GetValue()),
			Labels: labels,
			Unit:   ms.unitParser.Parse(valueMetric.GetUnit()),
		}}
	case events.Envelope_ContainerMetric:
		containerMetric := envelope.GetContainerMetric()
		metrics = []stackdriver.Metric{
			{Name: "diskBytesQuota", Points: points(float64(containerMetric.GetDiskBytesQuota())), Labels: labels},
			{Name: "instanceIndex", Points: points(float64(containerMetric.GetInstanceIndex())), Labels: labels},
			{Name: "cpuPercentage", Points: points(float64(containerMetric.GetCpuPercentage())), Labels: labels},
			{Name: "diskBytes", Points: points(float64(containerMetric.GetDiskBytes())), Labels: labels},
			{Name: "memoryBytes", Points: points(float64(containerMetric.GetMemoryBytes())), Labels: labels},
			{Name: "memoryBytesQuota", Points: points(float64(containerMetric.GetMemoryBytesQuota())), Labels: labels},
		}
	case events.Envelope_CounterEvent:
		counterEvent := envelope.GetCounterEvent()
		metrics = []stackdriver.Metric{{
			Name:   counterEvent.GetName(),
			Points: points(float64(counterEvent.GetTotal())),
			Labels: labels,
		}}
	default:
		return fmt.Errorf("unknown event type: %v", envelope.EventType)
	}

	for _, metric := range metrics {
		ms.metricBuffer.PostMetric(&metric)
	}
	return nil
}
