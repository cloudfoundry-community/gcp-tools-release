package stackdriver_test

import (
	"errors"

	"github.com/cloudfoundry-community/gcp-tools-release/src/stackdriver-nozzle/mocks"
	"github.com/cloudfoundry-community/gcp-tools-release/src/stackdriver-nozzle/stackdriver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("MetricsBuffer", func() {
	var (
		subject       stackdriver.MetricsBuffer
		errs          <-chan error
		metricAdapter *mocks.MetricAdapter
	)

	BeforeEach(func() {
		metricAdapter = &mocks.MetricAdapter{}

		subject, errs = stackdriver.NewMetricsBuffer(5, metricAdapter)
	})

	It("acts as a passthrough with a buffer size of 1", func() {
		subject, errs = stackdriver.NewMetricsBuffer(1, metricAdapter)

		metric := &stackdriver.Metric{}
		subject.PostMetric(metric)

		Expect(metricAdapter.PostedMetrics).To(HaveLen(1))
		Expect(metricAdapter.PostedMetrics[0]).To(Equal(*metric))

		metric = &stackdriver.Metric{}
		subject.PostMetric(metric)

		Expect(metricAdapter.PostedMetrics).To(HaveLen(2))
		Expect(metricAdapter.PostedMetrics[1]).To(Equal(*metric))

		Consistently(errs).ShouldNot(Receive())
	})

	It("only sends after the buffer size is reached", func() {
		subject.PostMetric(&stackdriver.Metric{Name: "a"})
		subject.PostMetric(&stackdriver.Metric{Name: "b"})
		subject.PostMetric(&stackdriver.Metric{Name: "c"})
		subject.PostMetric(&stackdriver.Metric{Name: "d"})

		Expect(metricAdapter.PostedMetrics).To(BeEmpty())

		subject.PostMetric(&stackdriver.Metric{Name: "e"})
		Expect(metricAdapter.PostedMetrics).To(HaveLen(5))

		Consistently(errs).ShouldNot(Receive())
	})

	It("sends errors through the error channel", func() {
		subject, errs = stackdriver.NewMetricsBuffer(1, metricAdapter)

		expectedErr := errors.New("fail")
		metricAdapter.PostMetricError = expectedErr

		metric := &stackdriver.Metric{}
		subject.PostMetric(metric)

		Expect(metricAdapter.PostedMetrics).To(HaveLen(1))

		var err error
		Eventually(errs).Should(Receive(&err))
		Expect(err).To(Equal(expectedErr))
	})

	It("combines metrics with the same name and labels", func() {
		name := "neatoMetric"
		labels := map[string]string{"foo":"bar"}
		eventTime1 := time.Now()
		metric := &stackdriver.Metric{
			Name: name,
			Labels: labels,
			Points: map[time.Time]float64{eventTime1: 123.456},
		}
		subject.PostMetric(metric)

		eventTime2 := time.Now()
		metric = &stackdriver.Metric{
			Name: name,
			Labels: labels,
			Points: map[time.Time]float64{eventTime2: 456.789},
		}
		subject.PostMetric(metric)

		subject.PostMetric(&stackdriver.Metric{Name: "a"})
		subject.PostMetric(&stackdriver.Metric{Name: "b"})
		subject.PostMetric(&stackdriver.Metric{Name: "c"})

		Expect(metricAdapter.PostedMetrics).To(BeEmpty())

		subject.PostMetric(&stackdriver.Metric{Name: "d"})
		Expect(metricAdapter.PostedMetrics).To(HaveLen(5))
		Expect(metricAdapter.PostedMetrics[0]).To(Equal(stackdriver.Metric{
			Name: name,
			Labels: labels,
			Points: map[time.Time]float64{
				eventTime1: 123.456,
				eventTime2: 456.789,
			},
		}))
	})
})
