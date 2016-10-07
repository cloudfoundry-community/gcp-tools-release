package nozzle_test

import (
	"errors"

	"sync"

	"stackdriver-nozzle/nozzle"

	"github.com/cloudfoundry-community/firehose-to-syslog/caching"

	"stackdriver-nozzle/serializer"

	"github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"stackdriver-nozzle/mocks"
)

var _ = Describe("Nozzle", func() {
	var (
		sdClient *MockStackdriverClient
		subject  nozzle.Nozzle
	)

	BeforeEach(func() {
		sdClient = NewMockStackdriverClient()
		subject = nozzle.Nozzle{
			StackdriverClient: sdClient,
			Serializer:        serializer.NewSerializer(caching.NewCachingEmpty(), nil),
		}
	})

	It("handles HttpStartStop", func() {
		eventType := events.Envelope_HttpStartStop
		envelope := &events.Envelope{EventType: &eventType}

		subject.HandleEvent(envelope)

		postedLog := sdClient.postedLogs[0]
		Expect(postedLog.payload).To(Equal(envelope))
		Expect(postedLog.labels).To(Equal(map[string]string{
			"eventType": "HttpStartStop",
		}))
	})

	Context("metrics", func() {

		It("should post the value metric", func() {
			metricName := "memoryStats.lastGCPauseTimeNS"
			metricValue := float64(536182)
			metricType := events.Envelope_ValueMetric

			valueMetric := events.ValueMetric{
				Name:  &metricName,
				Value: &metricValue,
			}

			envelope := &events.Envelope{
				EventType:   &metricType,
				ValueMetric: &valueMetric,
			}

			err := subject.HandleEvent(envelope)
			Expect(err).To(BeNil())

			postedMetric := sdClient.postedMetrics[0]
			Expect(postedMetric.name).To(Equal(metricName))
			Expect(postedMetric.value).To(Equal(metricValue))
			Expect(postedMetric.labels).To(Equal(map[string]string{
				"eventType": "ValueMetric",
			}))
		})

		It("should post the container metrics", func() {
			diskBytesQuota := uint64(1073741824)
			instanceIndex := int32(0)
			cpuPercentage := 0.061651273460637
			diskBytes := uint64(164634624)
			memoryBytes := uint64(16601088)
			memoryBytesQuota := uint64(33554432)
			eventTime := int64(123)
			applicationId := "ee2aa52e-3c8a-4851-b505-0cb9fe24806e"

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
				Timestamp:       &eventTime,
			}

			err := subject.HandleEvent(envelope)
			Expect(err).To(BeNil())

			labels := map[string]string{
				"eventType":     "ContainerMetric",
				"applicationId": applicationId,
			}
			Expect(len(sdClient.postedMetrics)).To(Equal(6))
			Expect(sdClient.postedMetrics).To(ConsistOf(
				PostedMetric{"diskBytesQuota", float64(1073741824), eventTime, labels},
				PostedMetric{"instanceIndex", float64(0), eventTime, labels},
				PostedMetric{"cpuPercentage", 0.061651273460637, eventTime, labels},
				PostedMetric{"diskBytes", float64(164634624), eventTime, labels},
				PostedMetric{"memoryBytes", float64(16601088), eventTime, labels},
				PostedMetric{"memoryBytesQuota", float64(33554432), eventTime, labels},
			))
		})

		It("returns error if client errors out", func() {
			sdClient.postMetricError = errors.New("fail")
			metricType := events.Envelope_ContainerMetric
			envelope := &events.Envelope{
				EventType:   &metricType,
				ValueMetric: nil,
			}

			err := subject.HandleEvent(envelope)

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("name: diskBytesQuota value: 0.000000, error: fail"))
			Expect(err.Error()).To(ContainSubstring("name: memoryBytesQuota value: 0.000000, error: fail"))
		})

		It("returns error if getting metric errors out", func() {
			const errMessage = "GetMetrics fail"
			mockSerializer := &mocks.MockSerializer{
				GetMetricsFn: func(*events.Envelope) ([]*serializer.Metric, error) {
					return nil, errors.New(errMessage)
				},
				IsLogFn: func(*events.Envelope) bool {
					return false
				},
			}
			subject = nozzle.Nozzle{
				StackdriverClient: nil,
				Serializer:        mockSerializer,
			}

			envelope := &events.Envelope{}

			err := subject.HandleEvent(envelope)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal(errMessage))
		})
	})
})

type MockStackdriverClient struct {
	postedLogs    []PostedLog
	postedMetrics []PostedMetric

	postMetricError error

	mutex *sync.Mutex
}

func NewMockStackdriverClient() *MockStackdriverClient {
	return &MockStackdriverClient{
		postedLogs:      []PostedLog{},
		postedMetrics:   []PostedMetric{},
		postMetricError: nil,
		mutex:           &sync.Mutex{},
	}
}

func (m *MockStackdriverClient) PostLog(payload interface{}, labels map[string]string) {
	m.mutex.Lock()
	m.postedLogs = append(m.postedLogs, PostedLog{payload, labels})
	m.mutex.Unlock()
}

func (m *MockStackdriverClient) PostMetric(name string, value float64, eventTime int64, labels map[string]string) error {
	m.mutex.Lock()
	m.postedMetrics = append(m.postedMetrics, PostedMetric{name, value, eventTime, labels})
	m.mutex.Unlock()

	return m.postMetricError
}

type PostedLog struct {
	payload interface{}
	labels  map[string]string
}

type PostedMetric struct {
	name      string
	value     float64
	eventTime int64
	labels    map[string]string
}
