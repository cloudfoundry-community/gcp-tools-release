package nozzle_test

import (
	"time"

	"github.com/evandbrown/gcp-tools-release/src/stackdriver-nozzle/nozzle"

	"github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Envelope", func() {
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

		envelope := nozzle.Envelope{
			Envelope: &events.Envelope{
				Origin:     &origin,
				EventType:  &eventType,
				Timestamp:  &timestamp,
				Deployment: &deployment,
				Job:        &job,
				Index:      &index,
				Ip:         &ip,
				Tags:       tags,
			},
		}

		labels := envelope.Labels()
		Expect(labels).To(Equal(map[string]string{
			"origin":     origin,
			"event_type": eventType.String(),
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

		envelope := nozzle.Envelope{
			Envelope: &events.Envelope{
				Origin:     &origin,
				EventType:  &eventType,
				Timestamp:  &timestamp,
				Deployment: nil,
				Job:        &job,
				Index:      &index,
				Ip:         nil,
				Tags:       tags,
			},
		}

		labels := envelope.Labels()
		Expect(labels).To(Equal(map[string]string{
			"origin":     origin,
			"event_type": eventType.String(),
			"job":        job,
			"index":      index,
		}))
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
				envelope := nozzle.Envelope{
					Envelope: &events.Envelope{
						EventType:     &eventType,
						HttpStartStop: &event,
					},
				}

				labels := envelope.Labels()
				Expect(labels["application_id"]).To(Equal(guid))
			})

			It("LogMessage adds app id", func() {
				eventType := events.Envelope_LogMessage

				event := events.LogMessage{
					AppId: &guid,
				}
				envelope := nozzle.Envelope{
					Envelope: &events.Envelope{
						EventType:  &eventType,
						LogMessage: &event,
					},
				}

				labels := envelope.Labels()
				Expect(labels["application_id"]).To(Equal(guid))

			})

			It("ValueMetric does not add app id", func() {
				eventType := events.Envelope_ValueMetric

				event := events.ValueMetric{}
				envelope := nozzle.Envelope{
					Envelope: &events.Envelope{
						EventType:   &eventType,
						ValueMetric: &event,
					},
				}

				labels := envelope.Labels()
				Expect(labels).NotTo(HaveKey("application_id"))

			})

			It("CounterEvent does not add app id", func() {
				eventType := events.Envelope_CounterEvent

				event := events.CounterEvent{}
				envelope := nozzle.Envelope{
					Envelope: &events.Envelope{
						EventType:    &eventType,
						CounterEvent: &event,
					},
				}

				labels := envelope.Labels()
				Expect(labels).NotTo(HaveKey("application_id"))

			})

			It("Error does not add app id", func() {
				eventType := events.Envelope_Error

				event := events.Error{}
				envelope := nozzle.Envelope{
					Envelope: &events.Envelope{
						EventType: &eventType,
						Error:     &event,
					},
				}

				labels := envelope.Labels()
				Expect(labels).NotTo(HaveKey("application_id"))

			})

			It("ContainerMetric does add app id", func() {
				eventType := events.Envelope_ContainerMetric

				event := events.ContainerMetric{
					ApplicationId: &guid,
				}
				envelope := nozzle.Envelope{
					Envelope: &events.Envelope{
						EventType:       &eventType,
						ContainerMetric: &event,
					},
				}

				labels := envelope.Labels()
				Expect(labels["application_id"]).To(Equal(guid))
			})
		})
	})
})
