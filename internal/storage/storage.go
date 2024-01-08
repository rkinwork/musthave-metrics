package storage

type MetricStorage interface {
	Get(m Metrics) (Metrics, bool)
	Set(m Metrics) error
	Delete(m Metrics) error
	IterMetrics() []Metrics
}
