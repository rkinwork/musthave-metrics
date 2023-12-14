package storage

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetFromRepository(t *testing.T) {
	type want struct {
		value string
		ok    bool
	}
	tests := []struct {
		name   string
		input  Metric
		metric Metric
		want   want
	}{
		{
			name: "get Gauge Value",
			input: Gauge{
				Name:  "test",
				Value: 99,
			},
			metric: Gauge{
				Name:  "test",
				Value: 0,
			},
			want: want{"99", true},
		},
		{
			name: "get Gauge with decimals Value",
			input: Gauge{
				Name:  "test",
				Value: 99.999,
			},
			metric: Gauge{
				Name:  "test",
				Value: 0,
			},
			want: want{"99.999", true},
		},
		{
			name: "get Counter Value",
			input: Counter{
				Name:  "test",
				Value: 99,
			},
			metric: Counter{
				Name:  "test",
				Value: 0,
			},
			want: want{"99", true},
		},
		{
			name: "get Counter Value absent",
			input: Counter{
				Name:  "test",
				Value: 99,
			},
			metric: Counter{
				Name:  "absent",
				Value: 0,
			},
			want: want{"0", false},
		},
		{
			name: "get Gauge Value absent",
			input: Gauge{
				Name:  "test",
				Value: 99,
			},
			metric: Gauge{
				Name:  "absent",
				Value: 0,
			},
			want: want{"0", false},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewInMemMetricStorage()
			c := NewInMemMetricStorage()
			repo := NewMetricRepository(g, c)
			var err error
			switch v := tc.input.(type) {
			case Gauge:
				err = g.Set(v)
			case Counter:
				err = c.Set(v)
			default:
				panic("unknown type")
			}
			require.NoError(t, err)
			metric, ok, err := repo.Get(tc.metric.ExportTypeName(), tc.metric.GetName())
			require.NoError(t, err)
			assert.Equal(t, tc.want.ok, ok)
			if ok {
				assert.Equal(t, tc.want.value, metric.ExportValue())
			}

		})
	}
}
