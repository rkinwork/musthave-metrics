package agent

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"math/rand"
	"runtime"
)

const PollCount = `PollCount`
const RandomValue = `RandomValue`
const retryCount = 3

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
		PollCount:       {storage.CounterMetric, `1`},
		RandomValue:     {storage.GaugeMetric, fmt.Sprintf(`%f'`, rand.Float64())},
		`Alloc`:         {storage.GaugeMetric, fmt.Sprintf(`%d`, m.Alloc)},
		`BuckHashSys`:   {storage.GaugeMetric, fmt.Sprintf(`%d`, m.BuckHashSys)},
		`Frees`:         {storage.GaugeMetric, fmt.Sprintf(`%d`, m.Frees)},
		`GCCPUFraction`: {storage.GaugeMetric, fmt.Sprintf(`%f`, m.GCCPUFraction)},
		`GCSys`:         {storage.GaugeMetric, fmt.Sprintf(`%d`, m.GCSys)},
		`HeapAlloc`:     {storage.GaugeMetric, fmt.Sprintf(`%d`, m.HeapAlloc)},
		`HeapIdle`:      {storage.GaugeMetric, fmt.Sprintf(`%d`, m.HeapIdle)},
		`HeapInuse`:     {storage.GaugeMetric, fmt.Sprintf(`%d`, m.HeapInuse)},
		`HeapObjects`:   {storage.GaugeMetric, fmt.Sprintf(`%d`, m.HeapObjects)},
		`HeapReleased`:  {storage.GaugeMetric, fmt.Sprintf(`%d`, m.HeapReleased)},
		`HeapSys`:       {storage.GaugeMetric, fmt.Sprintf(`%d`, m.HeapSys)},
		`LastGC`:        {storage.GaugeMetric, fmt.Sprintf(`%d`, m.LastGC)},
		`Lookups`:       {storage.GaugeMetric, fmt.Sprintf(`%d`, m.Lookups)},
		`MCacheInuse`:   {storage.GaugeMetric, fmt.Sprintf(`%d`, m.MCacheInuse)},
		`MCacheSys`:     {storage.GaugeMetric, fmt.Sprintf(`%d`, m.MCacheSys)},
		`MSpanInuse`:    {storage.GaugeMetric, fmt.Sprintf(`%d`, m.MSpanInuse)},
		`MSpanSys`:      {storage.GaugeMetric, fmt.Sprintf(`%d`, m.MSpanSys)},
		`Mallocs`:       {storage.GaugeMetric, fmt.Sprintf(`%d`, m.Mallocs)},
		`NextGC`:        {storage.GaugeMetric, fmt.Sprintf(`%d`, m.NextGC)},
		`NumForcedGC`:   {storage.GaugeMetric, fmt.Sprintf(`%d`, m.NumForcedGC)},
		`NumGC`:         {storage.GaugeMetric, fmt.Sprintf(`%d`, m.NumGC)},
		`OtherSys`:      {storage.GaugeMetric, fmt.Sprintf(`%d`, m.OtherSys)},
		`PauseTotalNs`:  {storage.GaugeMetric, fmt.Sprintf(`%d`, m.PauseTotalNs)},
		`StackInuse`:    {storage.GaugeMetric, fmt.Sprintf(`%d`, m.StackInuse)},
		`StackSys`:      {storage.GaugeMetric, fmt.Sprintf(`%d`, m.StackSys)},
		`Sys`:           {storage.GaugeMetric, fmt.Sprintf(`%d`, m.Sys)},
		`TotalAlloc`:    {storage.GaugeMetric, fmt.Sprintf(`%d`, m.TotalAlloc)},
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
	c := resty.New()
	c.SetRetryCount(retryCount)
	_, err := c.R().SetHeader("Content-Type", "text/plain").Post(endPoint)
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
