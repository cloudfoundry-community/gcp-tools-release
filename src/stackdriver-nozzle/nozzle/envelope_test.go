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

	It("httpStartStop adds app id when present", func() {
		origin := "cool-origin"
		eventType := events.Envelope_HttpStartStop
		timestamp := time.Now().UnixNano()
		job := "some-job"
		index := "an-index"

		low := uint64(0x7243cc580bc17af4)
		high := uint64(0x79d4c3b2020e67a5)
		appId := events.UUID{Low: &low, High: &high}

		event := events.HttpStartStop{
			ApplicationId: &appId,
		}
		envelope := nozzle.Envelope{
			Envelope: &events.Envelope{
				Origin:        &origin,
				EventType:     &eventType,
				Timestamp:     &timestamp,
				Deployment:    nil,
				Job:           &job,
				Index:         &index,
				Ip:            nil,
				HttpStartStop: &event,
			},
		}

		labels := envelope.Labels()
		Expect(labels["application_id"]).To(Equal("f47ac10b-58cc-4372-a567-0e02b2c3d479"))
	})

	It("httpStartStop adds app meta data", func() {

	})

})
