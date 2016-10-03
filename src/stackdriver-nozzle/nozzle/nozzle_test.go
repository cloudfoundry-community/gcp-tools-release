package nozzle_test

import (
	"errors"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/nozzle"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nozzle", func() {

	var (
		sdClient *MockStackdriverClient
		subject  nozzle.Nozzle
	)

	BeforeEach(func() {
		sdClient = &MockStackdriverClient{}
		subject = nozzle.Nozzle{StackdriverClient: sdClient}
	})

	Context("logging", func() {
		var envelope *events.Envelope

		BeforeEach(func() {
			eventType := events.Envelope_HttpStartStop
			envelope = &events.Envelope{
				EventType: &eventType,
			}
		})

		It("ships something to the stackdriver client", func() {
			subject.HandleEvent(envelope)

			postedLog := sdClient.postedLogs[0]
			Expect(postedLog.payload).To(Equal(nozzle.Envelope{envelope}))
			Expect(postedLog.labels).To(Equal(map[string]string{
				"event_type": "HttpStartStop",
			}))
		})

		It("ships multiple events", func() {
			for i := 0; i < 10; i++ {
				subject.HandleEvent(envelope)
			}

			Expect(len(sdClient.postedLogs)).To(Equal(10))
		})
	})

	Context("metrics", func() {
		var envelope *events.Envelope

		It("should post the value metric", func() {
			metricName := "memoryStats.lastGCPauseTimeNS"
			metricValue := float64(536182)
			metricType := events.Envelope_ValueMetric

			valueMetric := events.ValueMetric{
				Name:  &metricName,
				Value: &metricValue,
			}

			envelope = &events.Envelope{
				EventType:   &metricType,
				ValueMetric: &valueMetric,
			}

			err := subject.HandleEvent(envelope)
			Expect(err).To(BeNil())

			postedMetric := sdClient.postedMetrics[0]
			Expect(postedMetric.name).To(Equal(metricName))
			Expect(postedMetric.value).To(Equal(metricValue))
			Expect(postedMetric.labels).To(Equal(map[string]string{
				"event_type": "ValueMetric",
			}))
		})

		It("returns error if client errors out", func() {
			sdClient.postMetricError = errors.New("fail")

			err := subject.HandleEvent(envelope)

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("fail"))
		})
	})
})

type MockStackdriverClient struct {
	postedLogs    []PostedLog
	postedMetrics []PostedMetric

	postMetricError error
}

func (m *MockStackdriverClient) PostLog(payload interface{}, labels map[string]string) {
	m.postedLogs = append(m.postedLogs, PostedLog{payload, labels})
}

func (m *MockStackdriverClient) PostMetric(name string, value float64, labels map[string]string) error {
	m.postedMetrics = append(m.postedMetrics, PostedMetric{name, value, labels})

	return m.postMetricError
}

type PostedLog struct {
	payload interface{}
	labels  map[string]string
}

type PostedMetric struct {
	name   string
	value  float64
	labels map[string]string
}
