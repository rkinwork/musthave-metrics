package storage

import (
	"fmt"
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

type GaugeStorage interface {
	Get(metric Gauge) (Gauge, bool)
	Set(metric Gauge) error
	Delete(metric Gauge) error
	IterMetrics() []Metric
}

type CounterStorage interface {
	Get(metric Counter) (Counter, bool)
	Set(metric Counter) error
	Delete(metric Counter) error
	IterMetrics() []Metric
}

type InMemCounterStorage struct {
	m map[string]Counter
}

func (i *InMemCounterStorage) Get(metric Counter) (Counter, bool) {
	res, ok := i.m[metric.GetName()]
	return res, ok
}

func (i *InMemCounterStorage) Set(metric Counter) error {
	i.m[metric.GetName()] = metric
	return nil
}

func (i *InMemCounterStorage) Delete(metric Counter) error {
	delete(i.m, metric.GetName())
	return nil
}

func (i *InMemCounterStorage) IterMetrics() []Metric {
	var res []Metric
	for _, metric := range i.m {
		res = append(res, metric)
	}
	return res
}

type InMemGaugeStorage struct {
	m map[string]Gauge
}

func (i *InMemGaugeStorage) Get(metric Gauge) (Gauge, bool) {
	res, ok := i.m[metric.GetName()]
	return res, ok
}

func (i *InMemGaugeStorage) Set(metric Gauge) error {
	i.m[metric.GetName()] = metric
	return nil
}

func (i *InMemGaugeStorage) Delete(metric Gauge) error {
	delete(i.m, metric.GetName())
	return nil
}

func (i *InMemGaugeStorage) IterMetrics() []Metric {
	var res []Metric
	for _, metric := range i.m {
		res = append(res, metric)
	}
	return res
}

type MetricRepository struct {
	gaugeStorage   GaugeStorage
	counterStorage CounterStorage
}

func (m *MetricRepository) Collect(metric Metric) error {
	switch v := metric.(type) {
	case Counter:
		return m.Add(v)
	case Gauge:
		return m.Set(v)
	}
	return fmt.Errorf("unknown type %t", metric)
}

func (m *MetricRepository) Add(metric Metric) error {
	switch v := metric.(type) {
	case Counter:
		old, _ := m.counterStorage.Get(v)
		v.Value = v.Value + old.Value
		return m.counterStorage.Set(v)
	case Gauge:
		old, _ := m.gaugeStorage.Get(v)
		v.Value = v.Value + old.Value
		return m.gaugeStorage.Set(v)
	}
	return fmt.Errorf("unknown type %t", metric)
}

func (m *MetricRepository) Set(metric Metric) error {
	switch v := metric.(type) {
	case Counter:
		return m.counterStorage.Set(v)
	case Gauge:
		return m.gaugeStorage.Set(v)
	}
	return fmt.Errorf("unknown type %t", metric)
}

func (m *MetricRepository) Get(metric Metric) (Metric, bool, error) {
	switch v := metric.(type) {
	case Counter:
		res, ok := m.counterStorage.Get(v)
		return res, ok, nil
	case Gauge:
		res, ok := m.gaugeStorage.Get(v)
		return res, ok, nil
	}
	return metric, false, fmt.Errorf("unknown type %t", metric)
}

func (m *MetricRepository) Delete(metric Metric) error {
	switch v := metric.(type) {
	case Counter:
		return m.counterStorage.Delete(v)
	case Gauge:
		return m.gaugeStorage.Delete(v)
	}
	return fmt.Errorf("unknown type %t", metric)
}

func (m *MetricRepository) IterMetrics() []Metric {
	allMetrics := make([]Metric, 0)
	for _, metricType := range m.GetTypes() {
		metrics, err := m.GetMetricsBy(metricType)
		if err != nil {
			panic(err)
		}
		allMetrics = append(allMetrics, metrics...)
	}
	return allMetrics
}

func (m *MetricRepository) GetTypes() []string {
	return []string{GaugeMetric, CounterMetric}
}

func (m *MetricRepository) GetMetricsBy(metricType string) ([]Metric, error) {
	switch metricType {
	case CounterMetric:
		return m.counterStorage.IterMetrics(), nil
	case GaugeMetric:
		return m.gaugeStorage.IterMetrics(), nil
	}
	return []Metric{}, fmt.Errorf("unknown metric type %s", metricType)
}

func NewMetricRepository(g GaugeStorage, c CounterStorage) *MetricRepository {
	return &MetricRepository{
		g,
		c,
	}
}

func NewInMemMetricRepository() *MetricRepository {
	g := &InMemGaugeStorage{m: make(map[string]Gauge)}
	c := &InMemCounterStorage{m: make(map[string]Counter)}
	return NewMetricRepository(g, c)
}
