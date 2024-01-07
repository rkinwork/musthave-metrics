package storage

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func NewCounterMetrics(id string, delta int64) Metrics {
	var d = &delta
	return Metrics{
		ID:    id,
		MType: CounterMetric,
		Delta: d,
	}
}

func NewGaugeMetrics(id string, value float64) Metrics {
	var v = &value
	return Metrics{
		ID:    id,
		MType: GaugeMetric,
		Value: v,
	}
}

func TestGetFromRepository(t *testing.T) {
	type want struct {
		value Metrics
		ok    bool
	}
	tests := []struct {
		name   string
		input  Metrics
		metric Metrics
		want   want
	}{
		{
			name:  "get Gauge Value",
			input: NewGaugeMetrics("test", 99),
			metric: Metrics{
				ID:    "test",
				MType: GaugeMetric,
			},
			want: want{NewGaugeMetrics("test", 99), true},
		},
		{
			name:  "get Gauge with decimals Value",
			input: NewGaugeMetrics("test", 99.999),
			metric: Metrics{
				ID:    "test",
				MType: GaugeMetric,
			},
			want: want{NewGaugeMetrics("test", 99.999), true},
		},
		{
			name:  "get Counter Value",
			input: NewCounterMetrics("test", 99),
			metric: Metrics{
				ID:    "test",
				MType: CounterMetric,
			},
			want: want{NewCounterMetrics("test", 99), true},
		},
		{
			name:  "get Counter Value absent",
			input: NewCounterMetrics("test", 99),
			metric: Metrics{
				ID:    "absent",
				MType: CounterMetric,
			},
			want: want{Metrics{}, false},
		},
		{
			name:  "get Gauge Value absent",
			input: NewGaugeMetrics("test", 99),
			metric: Metrics{
				ID:    "absent",
				MType: GaugeMetric,
			},
			want: want{Metrics{}, false},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewInMemMetricRepository()
			var err error
			_, err = repo.Collect(tc.input)
			require.NoError(t, err)
			metric, ok := repo.Get(tc.metric)
			require.NoError(t, err)
			assert.Equal(t, tc.want.ok, ok)
			switch {
			case !ok:
				return
			case metric.MType == CounterMetric:
				assert.Equal(t, *tc.want.value.Delta, *metric.Delta)
			case metric.MType == GaugeMetric:
				assert.Equal(t, *tc.want.value.Value, *metric.Value)
			}

		})
	}
}
