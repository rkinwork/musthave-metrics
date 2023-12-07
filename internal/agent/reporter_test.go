package agent

import (
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCollectMemMetricsCounter(t *testing.T) {
	mStorage := storage.GetLocalStorageModel()
	knownMetrics := GetCollectdMetricStorage()
	CollectMemMetrics(mStorage, knownMetrics)
	val, err := mStorage.Get(storage.CounterMetric, PollCount)
	require.Nil(t, err)
	assert.Equal(t, `1`, val)
	CollectMemMetrics(mStorage, knownMetrics)
	val, err = mStorage.Get(storage.CounterMetric, PollCount)
	require.Nil(t, err)
	assert.Equal(t, `2`, val)
	CollectMemMetrics(mStorage, knownMetrics)
	val, err = mStorage.Get(storage.CounterMetric, PollCount)
	require.Nil(t, err)
	assert.NotEqual(t, `5`, val)

}
