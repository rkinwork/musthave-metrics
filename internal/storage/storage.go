package storage

import "sync"

type IMetricStorage interface {
	Get(m *Metrics) (Metrics, bool)
	Set(m Metrics) error
	Delete(m *Metrics) error
	IterMetrics() []Metrics
}

// In-memory storage

type InMemMetricStorage struct {
	m map[MetricHash]Metrics
	sync.Mutex
}

func (i *InMemMetricStorage) Get(m *Metrics) (Metrics, bool) {
	res, ok := i.m[m.GetHash()]
	return res, ok
}

func (i *InMemMetricStorage) Set(m Metrics) error {
	i.Lock()
	i.m[m.GetHash()] = m
	i.Unlock()
	return nil
}

func (i *InMemMetricStorage) Delete(m *Metrics) error {
	i.Lock()
	delete(i.m, m.GetHash())
	i.Unlock()
	return nil
}

func (i *InMemMetricStorage) IterMetrics() []Metrics {
	var res []Metrics
	for _, metric := range i.m {
		res = append(res, metric)
	}
	return res
}

func NewInMemMetricStorage() *InMemMetricStorage {
	imms := &InMemMetricStorage{
		m: make(map[MetricHash]Metrics),
	}

	return imms
}
