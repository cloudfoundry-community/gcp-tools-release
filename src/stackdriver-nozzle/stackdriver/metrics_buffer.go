package stackdriver

type MetricsBuffer interface {
	PostMetric(*Metric)
}

type metricsBuffer struct {
	size    int
	adapter MetricAdapter
	errs    chan error
	metrics []Metric
}

func NewMetricsBuffer(size int, adapter MetricAdapter) (MetricsBuffer, <-chan error) {
	errs := make(chan error)
	return &metricsBuffer{size, adapter, errs, []Metric{}}, errs
}

func (mb *metricsBuffer) PostMetric(metric *Metric) {
	mb.metrics = append(mb.metrics, *metric)
	if len(mb.metrics) < mb.size {
		return
	}

	err := mb.adapter.PostMetrics(mb.metrics)
	if err != nil {
		go func() { mb.errs <- err }()
	}

	mb.metrics = []Metric{}
}