package main

type remover interface {
	Delete(n string)
}

type adder interface {
	Add(n string, v int64)
}

type setter interface {
	Set(n string, v float64)
}

type Counter interface {
	adder
}

type Gauge interface {
	setter
}

type MemStorage struct {
	counter Counter
	gauge   Gauge
}

type LocalCounter struct {
	m map[string]int64
}

type LocalGauge struct {
	m map[string]float64
}

func (c *LocalCounter) Add(n string, v int64) {
	c.m[n] += v
}

func (g *LocalGauge) Set(n string, v float64) {
	g.m[n] = v
}

func InitLocalMemStorage() MemStorage {
	return struct {
		counter Counter
		gauge   Gauge
	}{
		counter: &LocalCounter{m: map[string]int64{}},
		gauge:   &LocalGauge{map[string]float64{}},
	}
}
