package agent

import (
	"github.com/go-resty/resty/v2"
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCollectMemMetricsCounter(t *testing.T) {
	repository := storage.NewInMemMetricRepository()
	CollectMemMetrics(repository)
	val, _, _ := repository.Get(storage.CounterMetric, PollCount)
	assert.Equal(t, `1`, val.ExportValue())
	CollectMemMetrics(repository)
	val, _, _ = repository.Get(storage.CounterMetric, PollCount)
	assert.Equal(t, `2`, val.ExportValue())
	CollectMemMetrics(repository)
	val, _, _ = repository.Get(storage.CounterMetric, PollCount)
	assert.NotEqual(t, `5`, val.ExportValue())

}

func TestNewMetricSender(t *testing.T) {
	tests := []struct {
		name           string
		serverAddress  string
		expectedPrefix string
	}{
		{
			name:           "Without prefix",
			serverAddress:  "localhost:9999",
			expectedPrefix: "http://localhost:9999",
		},
		{
			name:           "With IP",
			serverAddress:  "http://192.168.0.1:9999",
			expectedPrefix: "http://192.168.0.1:9999",
		},
		{
			name:           "With IP without protocol",
			serverAddress:  "192.168.0.1:9999",
			expectedPrefix: "http://192.168.0.1:9999",
		},
		{
			name:           "With colon at the start",
			serverAddress:  ":9999",
			expectedPrefix: "http://localhost:9999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := NewMetricSender(tt.serverAddress)

			assert.IsType(t, &MetricSender{}, sender)
			assert.IsType(t, &resty.Client{}, sender.Client)
			assert.Equal(t, tt.expectedPrefix, sender.ServerAddress)
		})
	}
}
