package stackdriver_test

import (
	"errors"

	"github.com/cloudfoundry-community/gcp-tools-release/src/stackdriver-nozzle/mocks"
	"github.com/cloudfoundry-community/gcp-tools-release/src/stackdriver-nozzle/stackdriver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		metric := &stackdriver.Metric{}
		subject.PostMetric(metric)
		subject.PostMetric(metric)
		subject.PostMetric(metric)
		subject.PostMetric(metric)

		Expect(metricAdapter.PostedMetrics).To(BeEmpty())

		subject.PostMetric(metric)
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
})
