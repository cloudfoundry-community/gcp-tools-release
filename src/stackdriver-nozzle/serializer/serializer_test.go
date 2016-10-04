package serializer_test

import (
	"time"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/serializer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Serializer", func() {
	var (
		subject serializer.Serializer
	)

	BeforeEach(func() {
		subject = serializer.NewSerializer(nil)
	})

	It("has labels equivalent to its fields", func() {
		origin := "cool-origin"
		eventType := events.Envelope_HttpStartStop
		timestamp := time.Now().UnixNano()
		deployment := "neat-deployment"
		job := "some-job"
		index := "an-index"
		ip := "192.168.1.1"
		tags := map[string]string{
			"foo": "bar",
		}

		envelope := &events.Envelope{
			Origin:     &origin,
			EventType:  &eventType,
			Timestamp:  &timestamp,
			Deployment: &deployment,
			Job:        &job,
			Index:      &index,
			Ip:         &ip,
			Tags:       tags,
		}

		log := subject.GetLog(envelope)

		labels := log.Labels
		Expect(labels).To(Equal(map[string]string{
			"origin":     origin,
			"eventType":  eventType.String(),
			"deployment": deployment,
			"job":        job,
			"index":      index,
			"ip":         ip,
		}))
	})

	It("ignores empty fields", func() {
		origin := "cool-origin"
		eventType := events.Envelope_HttpStartStop
		timestamp := time.Now().UnixNano()
		job := "some-job"
		index := "an-index"
		tags := map[string]string{
			"foo": "bar",
		}

		envelope := &events.Envelope{
			Origin:     &origin,
			EventType:  &eventType,
			Timestamp:  &timestamp,
			Deployment: nil,
			Job:        &job,
			Index:      &index,
			Ip:         nil,
			Tags:       tags,
		}

		log := subject.GetLog(envelope)
		labels := log.Labels

		Expect(labels).To(Equal(map[string]string{
			"origin":    origin,
			"eventType": eventType.String(),
			"job":       job,
			"index":     index,
		}))
	})

	Context("GetMetrics", func() {
		It("creates the proper metrics for ContainerMetric", func() {
			diskBytesQuota := uint64(1073741824)
			instanceIndex := int32(0)
			cpuPercentage := 0.061651273460637
			diskBytes := uint64(164634624)
			memoryBytes := uint64(16601088)
			memoryBytesQuota := uint64(33554432)
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
			}

			labels := map[string]string{
				"eventType":     "ContainerMetric",
				"applicationId": applicationId,
			}

			metrics := subject.GetMetrics(envelope)

			Expect(metrics).To(HaveLen(6))

			Expect(metrics).To(ContainElement(&serializer.Metric{"diskBytesQuota", float64(1073741824), labels}))
			Expect(metrics).To(ContainElement(&serializer.Metric{"instanceIndex", float64(0), labels}))
			Expect(metrics).To(ContainElement(&serializer.Metric{"cpuPercentage", 0.061651273460637, labels}))
			Expect(metrics).To(ContainElement(&serializer.Metric{"diskBytes", float64(164634624), labels}))
			Expect(metrics).To(ContainElement(&serializer.Metric{"memoryBytes", float64(16601088), labels}))
			Expect(metrics).To(ContainElement(&serializer.Metric{"memoryBytesQuota", float64(33554432), labels}))
		})
	})

	Context("isLog", func() {
		It("HttpStartStop is log", func() {
			eventType := events.Envelope_HttpStartStop

			envelope := &events.Envelope{
				EventType: &eventType,
			}
			Expect(subject.IsLog(envelope)).To(BeTrue())
		})

		It("LogMessage is log", func() {
			eventType := events.Envelope_LogMessage

			envelope := &events.Envelope{
				EventType: &eventType,
			}
			Expect(subject.IsLog(envelope)).To(BeTrue())
		})

		It("ValueMetric is *NOT* log", func() {
			eventType := events.Envelope_ValueMetric

			envelope := &events.Envelope{
				EventType: &eventType,
			}
			Expect(subject.IsLog(envelope)).To(BeFalse())
		})

		XIt("CounterEvent is *NOT* log", func() {
			eventType := events.Envelope_CounterEvent

			envelope := &events.Envelope{
				EventType: &eventType,
			}
			Expect(subject.IsLog(envelope)).To(BeFalse())

		})

		It("Error is log", func() {
			eventType := events.Envelope_Error

			envelope := &events.Envelope{
				EventType: &eventType,
			}
			Expect(subject.IsLog(envelope)).To(BeTrue())

		})

		It("ContainerMetric is *NOT* log", func() {
			eventType := events.Envelope_ContainerMetric

			envelope := &events.Envelope{
				EventType: &eventType,
			}
			Expect(subject.IsLog(envelope)).To(BeFalse())

		})

	})

	Context("Metadata", func() {

		var (
			guid  = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
			low   = uint64(0x7243cc580bc17af4)
			high  = uint64(0x79d4c3b2020e67a5)
			appId = events.UUID{Low: &low, High: &high}
		)

		Context("application id", func() {
			It("httpStartStop adds app id when present", func() {
				eventType := events.Envelope_HttpStartStop

				event := events.HttpStartStop{
					ApplicationId: &appId,
				}
				envelope := &events.Envelope{
					EventType:     &eventType,
					HttpStartStop: &event,
				}

				log := subject.GetLog(envelope)
				labels := log.Labels

				Expect(labels["applicationId"]).To(Equal(guid))
			})

			It("LogMessage adds app id", func() {
				eventType := events.Envelope_LogMessage

				event := events.LogMessage{
					AppId: &guid,
				}
				envelope := &events.Envelope{
					EventType:  &eventType,
					LogMessage: &event,
				}

				log := subject.GetLog(envelope)
				labels := log.Labels
				Expect(labels["applicationId"]).To(Equal(guid))

			})

			It("ValueMetric does not add app id", func() {
				eventType := events.Envelope_ValueMetric

				event := events.ValueMetric{}
				envelope := &events.Envelope{
					EventType:   &eventType,
					ValueMetric: &event,
				}
				metrics := subject.GetMetrics(envelope)

				Expect(metrics).To(HaveLen(1))
				valueMetric := metrics[0]

				labels := valueMetric.Labels
				Expect(labels).NotTo(HaveKey("applicationId"))

			})

			XIt("CounterEvent does not add app id", func() {
				//TODO
			})

			It("Error does not add app id", func() {
				eventType := events.Envelope_Error

				event := events.Error{}
				envelope := &events.Envelope{
					EventType: &eventType,
					Error:     &event,
				}

				log := subject.GetLog(envelope)
				labels := log.Labels
				Expect(labels).NotTo(HaveKey("applicationId"))

			})

			It("ContainerMetric does add app id", func() {
				eventType := events.Envelope_ContainerMetric

				event := events.ContainerMetric{
					ApplicationId: &guid,
				}
				envelope := &events.Envelope{
					EventType:       &eventType,
					ContainerMetric: &event,
				}

				metrics := subject.GetMetrics(envelope)

				Expect(len(metrics)).To(Not(Equal(0)))

				for _, metric := range metrics {
					labels := metric.Labels
					Expect(labels["applicationId"]).To(Equal(guid))

				}
			})
		})
	})
})
