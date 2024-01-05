package storage

type InMemMetricStorage struct {
	m map[MetricHash]Metrics
}

func (i *InMemMetricStorage) Get(m Metrics) (Metrics, bool) {
	res, ok := i.m[m.GetHash()]
	return res, ok
}

func (i *InMemMetricStorage) Set(m Metrics) error {
	i.m[m.GetHash()] = m
	return nil
}

func (i *InMemMetricStorage) Delete(m Metrics) error {
	delete(i.m, m.GetHash())
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
	return &InMemMetricStorage{m: make(map[MetricHash]Metrics)}
}
