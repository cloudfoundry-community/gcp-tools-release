package nozzle_test

import (
	"time"

	"github.com/cloudfoundry-community/gcp-tools-release/src/stackdriver-nozzle/mocks"
	"github.com/cloudfoundry-community/gcp-tools-release/src/stackdriver-nozzle/nozzle"
	"github.com/cloudfoundry-community/gcp-tools-release/src/stackdriver-nozzle/stackdriver"
	"github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type mockUnitParser struct {
	lastInput string
}

func (m *mockUnitParser) Parse(unit string) string {
	m.lastInput = unit
	return "{foo}"
}

var _ = Describe("MetricSink", func() {
	var (
		subject      nozzle.Sink
		metricBuffer *mocks.MetricsBuffer
		unitParser   *mockUnitParser
		labels       map[string]string
		points       func(time.Time, float64) map[time.Time]float64
	)

	BeforeEach(func() {
		labels = map[string]string{"foo": "bar"}
		labelMaker := &mocks.LabelMaker{Labels: labels}
		metricBuffer = &mocks.MetricsBuffer{}
		unitParser = &mockUnitParser{}

		subject = nozzle.NewMetricSink(labelMaker, metricBuffer, unitParser)

		points = func(eventTime time.Time, value float64) map[time.Time]float64 {
			return map[time.Time]float64{eventTime: value}
		}
	})

	It("creates metric for ValueMetric", func() {
		eventTime := time.Now()

		name := "valueMetricName"
		value := 123.456
		unit := "barUnit"
		event := events.ValueMetric{
			Name:  &name,
			Value: &value,
			Unit:  &unit,
		}

		eventType := events.Envelope_ValueMetric
		timeStamp := eventTime.UnixNano()
		envelope := &events.Envelope{
			EventType:   &eventType,
			ValueMetric: &event,
			Timestamp:   &timeStamp,
		}

		err := subject.Receive(envelope)
		Expect(err).To(BeNil())

		metrics := metricBuffer.PostedMetrics
		Expect(metrics).To(ConsistOf(stackdriver.Metric{
			"valueMetricName",
			labels,
			points(eventTime, 123.456),
			"{foo}",
		}))

		Expect(unitParser.lastInput).To(Equal("barUnit"))
	})

	It("creates the proper metrics for ContainerMetric", func() {
		eventTime := time.Now()

		diskBytesQuota := uint64(1073741824)
		instanceIndex := int32(0)
		cpuPercentage := 0.061651273460637
		diskBytes := uint64(164634624)
		memoryBytes := uint64(16601088)
		memoryBytesQuota := uint64(33554432)
		applicationId := "ee2aa52e-3c8a-4851-b505-0cb9fe24806e"
		timeStamp := eventTime.UnixNano()

		metricType := events.Envelope_ContainerMetric
		containerMetric := events.ContainerMetric{
			DiskBytesQuota:   &diskBytesQuota,
			InstanceIndex:    &instanceIndex,
			CpuPercentage:    &cpuPercentage,
			DiskBytes:        &diskBytes,
			MemoryBytes:      &memoryBytes,
			MemoryBytesQuota: &memoryBytesQuota,
			ApplicationId:    &applicationId,
		}

		envelope := &events.Envelope{
			EventType:       &metricType,
			ContainerMetric: &containerMetric,
			Timestamp:       &timeStamp,
		}

		err := subject.Receive(envelope)
		Expect(err).To(BeNil())

		metrics := metricBuffer.PostedMetrics
		Expect(metrics).To(HaveLen(6))

		Expect(metrics).To(ContainElement(stackdriver.Metric{"diskBytesQuota", labels, points(eventTime, float64(1073741824)), ""}))
		Expect(metrics).To(ContainElement(stackdriver.Metric{"instanceIndex", labels, points(eventTime, float64(0)), ""}))
		Expect(metrics).To(ContainElement(stackdriver.Metric{"cpuPercentage", labels, points(eventTime, 0.061651273460637), ""}))
		Expect(metrics).To(ContainElement(stackdriver.Metric{"diskBytes", labels, points(eventTime, float64(164634624)), ""}))
		Expect(metrics).To(ContainElement(stackdriver.Metric{"memoryBytes", labels, points(eventTime, float64(16601088)), ""}))
		Expect(metrics).To(ContainElement(stackdriver.Metric{"memoryBytesQuota", labels, points(eventTime, float64(33554432)), ""}))
	})

	It("creates metric for CounterEvent", func() {
		eventTime := time.Now()

		eventType := events.Envelope_CounterEvent
		name := "counterName"
		total := uint64(123456)
		timeStamp := eventTime.UnixNano()

		event := events.CounterEvent{
			Name:  &name,
			Total: &total,
		}
		envelope := &events.Envelope{
			EventType:    &eventType,
			CounterEvent: &event,
			Timestamp:    &timeStamp,
		}

		err := subject.Receive(envelope)
		Expect(err).To(BeNil())

		metrics := metricBuffer.PostedMetrics
		Expect(metrics).To(ConsistOf(stackdriver.Metric{
			"counterName",
			labels,
			points(eventTime, float64(123456)),
			"",
		}))
	})

	It("returns error when envelope contains unhandled event type", func() {
		eventType := events.Envelope_HttpStart
		envelope := &events.Envelope{
			EventType: &eventType,
		}

		err := subject.Receive(envelope)

		Expect(err).NotTo(BeNil())
	})
})
