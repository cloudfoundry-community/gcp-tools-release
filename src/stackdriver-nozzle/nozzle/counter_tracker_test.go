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

package nozzle

import (
	"context"
	"time"

	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CounterTracker", func() {
	var (
		subject    *counterTracker
		counterTTL time.Duration
		logger     *mocks.MockLogger
	)

	BeforeEach(func() {
		logger = &mocks.MockLogger{}
		counterTTL = time.Duration(50) * time.Millisecond
		countersExpiredCount.Set(0)
	})

	It("increments counters and handles counter resets", func() {
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		subject = NewCounterTracker(ctx, counterTTL, logger)

		baseTime := time.Now()

		incomingTotals := []float64{10, 15, 25, 40, 10, 20}
		expectedTotals := []float64{5, 15, 30, 40, 50}

		for idx, value := range incomingTotals {
			ts := baseTime.Add(time.Duration(idx) * time.Millisecond)
			total, st := subject.GetTotal("counterName", value, &ts)
			if idx == 0 {
				// First seen value initializes the counter.
				Expect(st).To(BeNil())
			} else {
				Expect(total).To(Equal(expectedTotals[idx-1]))
				Expect(*st).To(BeTemporally("~", baseTime))
			}
		}
	})

	It("expires old counters", func() {
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		subject = NewCounterTracker(ctx, counterTTL, logger)

		runTest := func(baseTime time.Time) {
			incomingTotals := []float64{150, 165, 165, 170, 200, 200}
			expectedTotals := []float64{15, 15, 20, 50, 50}

			for idx, value := range incomingTotals {
				ts := baseTime.Add(time.Duration(idx) * time.Millisecond)
				total, st := subject.GetTotal("counterName2", value, &ts)
				if idx == 0 {
					// First seen value initializes the counter.
					Expect(st).To(BeNil())
				} else {
					Expect(total).To(Equal(expectedTotals[idx-1]))
					Expect(*st).To(BeTemporally("~", baseTime))
				}
			}
		}
		runTest(time.Now())
		Eventually(countersExpiredCount.IntValue).Should(Equal(1))
		runTest(time.Now())
	})
})
