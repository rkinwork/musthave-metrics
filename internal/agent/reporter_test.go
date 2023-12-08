package agent

import (
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCollectMemMetricsCounter(t *testing.T) {
	mStorage := storage.GetLocalStorageModel()
	knownMetrics := GetCollectdMetricStorage()
	CollectMemMetrics(mStorage, knownMetrics)
	val, _ := mStorage.Get(storage.CounterMetric, PollCount)
	assert.Equal(t, `1`, val)
	CollectMemMetrics(mStorage, knownMetrics)
	val, _ = mStorage.Get(storage.CounterMetric, PollCount)
	assert.Equal(t, `2`, val)
	CollectMemMetrics(mStorage, knownMetrics)
	val, _ = mStorage.Get(storage.CounterMetric, PollCount)
	assert.NotEqual(t, `5`, val)

}
