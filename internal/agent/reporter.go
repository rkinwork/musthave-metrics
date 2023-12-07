package agent

import (
	"fmt"
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"math/rand"
	"net/http"
	"runtime"
)

const PollCount = `PollCount`
const RandomValue = `RandomValue`

type ExtractedMetric struct {
	mType  string
	mValue string
}

type collectedMetric struct {
	name  string
	mType string
}
type CollectedMetrics struct {
	m map[collectedMetric]struct{}
}

func GetCollectdMetricStorage() *CollectedMetrics {
	return &CollectedMetrics{make(map[collectedMetric]struct{})}
}

func getMemMetrics() map[string]ExtractedMetric {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	return map[string]ExtractedMetric{
		PollCount:     {storage.CounterMetric, `1`},
		RandomValue:   {storage.GaugeMetric, fmt.Sprintf(`%f'`, rand.Float64())},
		`Alloc`:       {storage.GaugeMetric, fmt.Sprintf(`%d`, m.Alloc)},
		`BuckHashSys`: {storage.GaugeMetric, fmt.Sprintf(`%d`, m.BuckHashSys)},
	}
}

// CollectMemMetrics Every invocation adds metrics to the storage of metrics
func CollectMemMetrics(storage storage.MemStorageModelInt, metrics *CollectedMetrics) {
	mm := getMemMetrics()
	for mName, metric := range mm {
		_ = storage.InsertBy(metric.mType, mName, metric.mValue)
		metrics.m[collectedMetric{
			name:  mName,
			mType: metric.mType,
		}] = struct{}{}
	}
}

type MetricSender struct {
	ServerAddress string
}

func (m MetricSender) sendMetric(mType, mName, mVal string) error {
	endPoint := fmt.Sprintf(`%s/update/%s/%s/%s`, m.ServerAddress, mType, mName, mVal)
	//do not know what to do if request failed. just ignore
	resp, err := http.Post(endPoint, "Content-Type: text/plain", nil)
	if resp != nil {
		_ = resp.Body.Close()
	}
	return err
}

func SendMetrics(
	storage storage.MemStorageModelInt,
	metrics *CollectedMetrics,
	sender MetricSender,
) {
	for metric := range metrics.m {
		val, _ := storage.Get(metric.mType, metric.name)
		_ = sender.sendMetric(metric.mType, metric.name, val)
	}
}
