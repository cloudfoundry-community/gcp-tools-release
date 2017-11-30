/*
 * Copyright 2017 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package nozzle_test

import (
	"errors"

	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/mocks"
	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/nozzle"
	"github.com/cloudfoundry/lager"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nozzle", func() {
	var (
		subject    nozzle.Nozzle
		firehose   *mocks.FirehoseClient
		logSink    *mocks.NozzleSink
		metricSink *mocks.NozzleSink
		logger     *mocks.MockLogger
	)

	BeforeEach(func() {
		firehose = mocks.NewFirehoseClient()
		logSink = &mocks.NozzleSink{}
		metricSink = &mocks.NozzleSink{}
		logger = &mocks.MockLogger{}

		subject = nozzle.NewNozzle(logger, logSink, metricSink)
		subject.Start(firehose)
	})

	It("updates the counter", func() {
		nozzle.FirehoseEventsTotal.Set(0)
		nozzle.FirehoseEventsReceived.Set(0)

		for _, value := range events.Envelope_EventType_value {
			eventType := events.Envelope_EventType(value)
			event := events.Envelope{EventType: &eventType}
			firehose.Messages <- &event
		}

		count := len(events.Envelope_EventType_value)
		Eventually(func() int {
			return nozzle.FirehoseEventsReceived.IntValue()
		}).Should(Equal(count))
		Expect(nozzle.FirehoseEventsTotal.IntValue()).To(Equal(count))
	})

	It("does not receive errors", func() {
		for _, value := range events.Envelope_EventType_value {
			eventType := events.Envelope_EventType(value)
			event := events.Envelope{EventType: &eventType}
			firehose.Messages <- &event
		}
	})

	It("handles HttpStartStop event", func() {
		eventType := events.Envelope_HttpStartStop
		envelope := &events.Envelope{EventType: &eventType}

		firehose.Messages <- envelope

		Eventually(logSink.LastEnvelope).Should(Equal(envelope))
	})

	It("handles LogMessage event", func() {
		eventType := events.Envelope_LogMessage
		envelope := &events.Envelope{EventType: &eventType}

		firehose.Messages <- envelope

		Eventually(logSink.LastEnvelope).Should(Equal(envelope))
	})

	It("handles Error event", func() {
		eventType := events.Envelope_Error
		envelope := &events.Envelope{EventType: &eventType}

		firehose.Messages <- envelope

		Eventually(logSink.LastEnvelope).Should(Equal(envelope))
	})

	It("handles ValueMetric event", func() {
		eventType := events.Envelope_ValueMetric
		envelope := &events.Envelope{EventType: &eventType}

		firehose.Messages <- envelope

		Eventually(metricSink.LastEnvelope).Should(Equal(envelope))
	})

	It("handles ContainerMetric event", func() {
		eventType := events.Envelope_ContainerMetric
		envelope := &events.Envelope{EventType: &eventType}

		firehose.Messages <- envelope

		Eventually(metricSink.LastEnvelope).Should(Equal(envelope))
	})

	It("handles CounterEvent event", func() {
		eventType := events.Envelope_CounterEvent
		envelope := &events.Envelope{EventType: &eventType}

		firehose.Messages <- envelope

		Eventually(metricSink.LastEnvelope).Should(Equal(envelope))
	})

	It("logs the firehose client errors", func() {
		err := errors.New("omg")
		go func() { firehose.Errs <- err }()

		Eventually(logger.Logs).Should(ContainElement(mocks.Log{
			Level:  lager.ERROR,
			Err:    err,
			Action: "firehose",
		}))
	})

	It("crashes on unrecoverable firehose errors", func() {
		err := consumer.ErrMaxRetriesReached
		go func() { firehose.Errs <- err }()
		Eventually(logger.Logs).Should(ContainElement(mocks.Log{
			Level:  lager.FATAL,
			Err:    err,
			Action: "firehose",
		}))
	})

	It("is resilient to multiple exists", func(done Done) {
		Expect(subject.Stop()).NotTo(HaveOccurred())
		Expect(subject.Stop()).To(HaveOccurred())
		Expect(subject.Stop()).To(HaveOccurred())

		close(done)
	}, 0.2)
})
