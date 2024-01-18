package storage

type IMetricRepository interface {
	Get(metric *Metrics) (Metrics, bool)
	Collect(metric *Metrics) (*Metrics, error)
	Set(metric *Metrics) (*Metrics, error)
	Delete(metric *Metrics) error
	GetAllMetrics() []Metrics
	Ping() error
}

type MetricRepository struct {
	storage IMetricStorage
}

func (m *MetricRepository) Get(metric *Metrics) (Metrics, bool) {
	return m.storage.Get(metric)
}

func (m *MetricRepository) Collect(metric *Metrics) (*Metrics, error) {
	switch metric.MType {
	case CounterMetric:
		delta := *metric.Delta
		if oldMetric, ok := m.storage.Get(metric); ok {
			delta = *oldMetric.Delta + delta
		}
		metric.Delta = &delta
	}
	return metric, m.storage.Set(*metric)
}

func (m *MetricRepository) Set(metrics *Metrics) (*Metrics, error) {
	return metrics, m.storage.Set(*metrics)
}

func (m *MetricRepository) Delete(metric *Metrics) error {
	return m.storage.Delete(metric)
}

func (m *MetricRepository) GetAllMetrics() []Metrics {
	return m.storage.IterMetrics()
}

func (m *MetricRepository) Ping() error {
	return nil
}

func NewRepository() IMetricRepository {
	return &MetricRepository{storage: NewInMemMetricStorage()}
}
