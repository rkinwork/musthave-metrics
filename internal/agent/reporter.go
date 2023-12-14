package agent

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"log"
	"math/rand"
	"runtime"
	"strings"
)

const PollCount = `PollCount`
const retries = 3

func getMemMetrics() map[string]storage.Metric {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	// use map to find out duplicates in metrics
	return map[string]storage.Metric{
		PollCount:       storage.Counter{Name: PollCount, Value: 1},
		`RandomValue`:   storage.Gauge{Name: `RandomValue`, Value: rand.Float64()},
		`Alloc`:         storage.Gauge{Name: `Alloc`, Value: float64(m.Alloc)},
		`BuckHashSys`:   storage.Gauge{Name: `BuckHashSys`, Value: float64(m.BuckHashSys)},
		`Frees`:         storage.Gauge{Name: `Frees`, Value: float64(m.Frees)},
		`GCCPUFraction`: storage.Gauge{Name: `GCCPUFraction`, Value: m.GCCPUFraction},
		`GCSys`:         storage.Gauge{Name: `GCSys`, Value: float64(m.GCSys)},
		`HeapAlloc`:     storage.Gauge{Name: `HeapAlloc`, Value: float64(m.HeapAlloc)},
		`HeapIdle`:      storage.Gauge{Name: `HeapIdle`, Value: float64(m.HeapIdle)},
		`HeapInuse`:     storage.Gauge{Name: `HeapInuse`, Value: float64(m.HeapInuse)},
		`HeapObjects`:   storage.Gauge{Name: `HeapObjects`, Value: float64(m.HeapObjects)},
		`HeapReleased`:  storage.Gauge{Name: `HeapReleased`, Value: float64(m.HeapReleased)},
		`HeapSys`:       storage.Gauge{Name: `HeapSys`, Value: float64(m.HeapSys)},
		`LastGC`:        storage.Gauge{Name: `LastGC`, Value: float64(m.LastGC)},
		`Lookups`:       storage.Gauge{Name: `Lookups`, Value: float64(m.Lookups)},
		`MCacheInuse`:   storage.Gauge{Name: `MCacheInuse`, Value: float64(m.MCacheInuse)},
		`MCacheSys`:     storage.Gauge{Name: `MCacheSys`, Value: float64(m.MCacheSys)},
		`MSpanInuse`:    storage.Gauge{Name: `MSpanInuse`, Value: float64(m.MSpanInuse)},
		`MSpanSys`:      storage.Gauge{Name: `MSpanSys`, Value: float64(m.MSpanSys)},
		`Mallocs`:       storage.Gauge{Name: `Mallocs`, Value: float64(m.Mallocs)},
		`NextGC`:        storage.Gauge{Name: `NextGC`, Value: float64(m.NextGC)},
		`NumForcedGC`:   storage.Gauge{Name: `NumForcedGC`, Value: float64(m.NumForcedGC)},
		`NumGC`:         storage.Gauge{Name: `NumGC`, Value: float64(m.NumGC)},
		`OtherSys`:      storage.Gauge{Name: `OtherSys`, Value: float64(m.OtherSys)},
		`PauseTotalNs`:  storage.Gauge{Name: `PauseTotalNs`, Value: float64(m.PauseTotalNs)},
		`StackInuse`:    storage.Gauge{Name: `StackInuse`, Value: float64(m.StackInuse)},
		`StackSys`:      storage.Gauge{Name: `StackSys`, Value: float64(m.StackSys)},
		`Sys`:           storage.Gauge{Name: `Sys`, Value: float64(m.Sys)},
		`TotalAlloc`:    storage.Gauge{Name: `TotalAlloc`, Value: float64(m.TotalAlloc)},
	}
}

// CollectMemMetrics Every invocation adds metrics to the storage of metrics
func CollectMemMetrics(repository *storage.MetricRepository) {
	mm := getMemMetrics()
	for _, metric := range mm {
		if err := repository.Collect(metric); err != nil {
			log.Printf("Problems with saving metric %v", metric)
		}
	}
}

type MetricSender struct {
	ServerAddress string
	*resty.Client
}

func (s *MetricSender) SendMetric(metric storage.Metric) error {
	endPoint := fmt.Sprintf(`%s/update/%s/%s/%s`, s.ServerAddress, metric.ExportTypeName(), metric.GetName(), metric.ExportValue())
	_, err := s.R().SetHeader("Content-Type", "text/plain").Post(endPoint)
	return err
}

func NewMetricSender(serverAddress string) *MetricSender {
	if strings.HasPrefix(serverAddress, `:`) {
		serverAddress = `localhost` + serverAddress
	}
	if !strings.HasPrefix(serverAddress, `http://`) {
		serverAddress = `http://` + serverAddress
	}
	c := resty.New()
	c.SetRetryCount(retries)
	return &MetricSender{
		ServerAddress: serverAddress,
		Client:        c,
	}
}

func SendMetrics(
	storage *storage.MetricRepository,
	sender *MetricSender,
) {

	for _, metric := range storage.IterMetrics() {
		if err := sender.SendMetric(metric); err != nil {
			log.Printf("problems with sending: %v, %e", metric, err)
		}

	}
}
