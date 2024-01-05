package storage

type MetricStorage interface {
	Get(metric Metrics) (Metrics, bool)
	Set(metric Metrics) error
	Delete(metric Metrics) error
	IterMetrics() []Metrics
}
