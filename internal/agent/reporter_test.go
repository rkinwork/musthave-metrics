package agent

import (
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCollectMemMetricsCounter(t *testing.T) {
	repository := storage.NewInMemMetricRepository()
	CollectMemMetrics(repository)
	val, _, _ := repository.Get(storage.Counter{Name: PollCount})
	assert.Equal(t, `1`, val.ExportValue())
	CollectMemMetrics(repository)
	val, _, _ = repository.Get(storage.Counter{Name: PollCount})
	assert.Equal(t, `2`, val.ExportValue())
	CollectMemMetrics(repository)
	val, _, _ = repository.Get(storage.Counter{Name: PollCount})
	assert.NotEqual(t, `5`, val.ExportValue())

}
