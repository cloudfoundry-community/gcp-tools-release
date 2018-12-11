package metricspipeline

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMetricsBuffer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MetricsPipeline Suite")
}
