package storage

import (
	"fmt"
	"github.com/rkinwork/musthave-metrics/internal/config"
	"strconv"
)

const (
	GaugeMetric   = `gauge`
	CounterMetric = `counter`
)

type Metric interface {
	GetName() string
	ExportValue() string
	ExportTypeName() string
}

type Counter struct {
	Name  string
	Value int64
}

func (c Counter) GetName() string {
	return c.Name
}

func (c Counter) ExportValue() string {
	return fmt.Sprintf("%d", c.Value)
}

func (c Counter) ExportTypeName() string {
	return CounterMetric
}

type Gauge struct {
	Name  string
	Value float64
}

func (g Gauge) GetName() string {
	return g.Name
}

func (g Gauge) ExportValue() string {
	return strconv.FormatFloat(g.Value, 'f', -1, 64)
}

func (g Gauge) ExportTypeName() string {
	return GaugeMetric
}

type MetricStorage interface {
	Get(name string) (Metric, bool)
	Set(metric Metric) error
	Delete(name string) error
	IterMetrics() []Metric
}

type InMemMetricStorage struct {
	m map[string]Metric
}

func (i *InMemMetricStorage) Get(name string) (Metric, bool) {
	res, ok := i.m[name]
	return res, ok
}

func (i *InMemMetricStorage) Set(metric Metric) error {
	i.m[metric.GetName()] = metric
	return nil
}

func (i *InMemMetricStorage) Delete(name string) error {
	delete(i.m, name)
	return nil
}

func (i *InMemMetricStorage) IterMetrics() []Metric {
	var res []Metric
	for _, metric := range i.m {
		res = append(res, metric)
	}
	return res
}

type MetricRepository struct {
	storage map[string]MetricStorage
}

func (m *MetricRepository) Get(metricType, name string) (Metric, bool, error) {
	storage, ok := m.storage[metricType]
	if !ok {
		return nil, false, fmt.Errorf("invalid metric type: %s", metricType)
	}
	metric, found := storage.Get(name)
	return metric, found, nil
}

func (m *MetricRepository) GetStorage(metric Metric) (MetricStorage, error) {
	storage, ok := m.storage[metric.ExportTypeName()]
	if !ok {
		return nil, fmt.Errorf("invalid metric type: %s", metric.ExportTypeName())
	}
	return storage, nil
}

func (m *MetricRepository) Collect(metric Metric) error {
	storage, err := m.GetStorage(metric)
	if err != nil {
		return err
	}

	switch v := metric.(type) {
	case Counter:
		if oldMetric, ok := storage.Get(metric.GetName()); ok {
			v.Value += oldMetric.(Counter).Value
			return storage.Set(v)
		}
	}
	return storage.Set(metric)
}

func (m *MetricRepository) Delete(metric Metric) error {
	storage, err := m.GetStorage(metric)
	if err != nil {
		return err
	}
	return storage.Delete(metric.GetName())
}

func (m *MetricRepository) GetAllMetrics() []Metric {
	allMetrics := make([]Metric, 0)
	for _, storage := range m.storage {
		allMetrics = append(allMetrics, storage.IterMetrics()...)
	}
	return allMetrics
}

func NewMetricRepository(g MetricStorage, c MetricStorage) *MetricRepository {
	return &MetricRepository{
		storage: map[string]MetricStorage{
			GaugeMetric:   g,
			CounterMetric: c,
		},
	}
}

func NewRepository(cfg *config.Config) *MetricRepository {
	switch cfg.StorageType {
	default:
		return NewInMemMetricRepository()
	}
}

func NewInMemMetricStorage() *InMemMetricStorage {
	return &InMemMetricStorage{m: make(map[string]Metric)}
}

func NewInMemMetricRepository() *MetricRepository {
	g := NewInMemMetricStorage()
	c := NewInMemMetricStorage()
	return NewMetricRepository(g, c)
}
