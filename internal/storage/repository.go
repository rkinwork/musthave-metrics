package storage

import (
	"github.com/rkinwork/musthave-metrics/internal/config"
)

type IMetricRepository interface {
	Get(metric Metrics) (Metrics, bool, error)
	Collect(metric Metrics) (Metrics, error)
	Delete(metric Metrics) error
	GetAllMetrics() []Metrics
}

type MetricRepository struct {
	storage MetricStorage
}

func (m *MetricRepository) Get(metric Metrics) (Metrics, bool) {
	return m.storage.Get(metric)
}

func (m *MetricRepository) Collect(metric Metrics) (Metrics, error) {
	switch metric.MType {
	case CounterMetric:
		delta := *metric.Delta
		if oldMetric, ok := m.storage.Get(metric); ok {
			delta = *oldMetric.Delta + delta
		}
		metric.Delta = &delta
	}
	return metric, m.storage.Set(metric)
}

func (m *MetricRepository) Delete(metric Metrics) error {
	return m.storage.Delete(metric)
}

func (m *MetricRepository) GetAllMetrics() []Metrics {
	return m.storage.IterMetrics()
}

func NewRepository(cfg *config.Config) *MetricRepository {
	switch cfg.StorageType {
	default:
		return NewInMemMetricRepository()
	}
}

func NewInMemMetricRepository() *MetricRepository {
	return &MetricRepository{storage: NewInMemMetricStorage()}
}
