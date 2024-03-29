package storage

type MetricHash string

const (
	GaugeMetric   = `gauge`
	CounterMetric = `counter`
)

// Metrics is a struct that represents a metric.
// It contains the ID, which is the name of the metric.
// The MType parameter can have a value of "gauge" or "counter".
// If MType is "counter", the Delta field represents the value of the metric.
// If MType is "gauge", the Value field represents the value of the metric.
// If we use Metrics as response to request we fill only Value
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// GetHash returns a hash string composed of the ID and MType fields of the Metrics struct.
func (m Metrics) GetHash() MetricHash {
	return MetricHash(m.ID + m.MType)
}
