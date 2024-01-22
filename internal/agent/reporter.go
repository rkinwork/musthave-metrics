package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
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

type MemExtractor struct {
	ID          string
	MType       string
	ExtractorFn func(*runtime.MemStats) string
}

var presets = map[string]MemExtractor{
	PollCount: {
		ID:    PollCount,
		MType: storage.CounterMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return "1"
		},
	},
	`RandomValue`: {
		ID:    `RandomValue`,
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf(`%f`, rand.Float64())
		},
	},
	`Alloc`: {
		ID:    `Alloc`,
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf(`%d`, stats.Alloc)
		},
	},
	"BuckHashSys": {
		ID:    "BuckHashSys",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.BuckHashSys)
		},
	},
	"Frees": {
		ID:    "Frees",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.Frees)
		},
	},
	"GCCPUFraction": {
		ID:    "GCCPUFraction",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%f", stats.GCCPUFraction)
		},
	},
	"GCSys": {
		ID:    "GCSys",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.GCSys)
		},
	},
	"HeapAlloc": {
		ID:    "HeapAlloc",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.HeapAlloc)
		},
	},
	"HeapIdle": {
		ID:    "HeapIdle",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.HeapIdle)
		},
	},
	"HeapInuse": {
		ID:    "HeapInuse",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.HeapInuse)
		},
	},
	"HeapObjects": {
		ID:    "HeapObjects",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.HeapObjects)
		},
	},
	"HeapReleased": {
		ID:    "HeapReleased",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.HeapReleased)
		},
	},
	"HeapSys": {
		ID:    "HeapSys",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.HeapSys)
		},
	},
	"LastGC": {
		ID:    "LastGC",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.LastGC)
		},
	},
	"Lookups": {
		ID:    "Lookups",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.Lookups)
		},
	},
	"MCacheInuse": {
		ID:    "MCacheInuse",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.MCacheInuse)
		},
	},
	"MCacheSys": {
		ID:    "MCacheSys",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.MCacheSys)
		},
	},
	"MSpanInuse": {
		ID:    "MSpanInuse",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.MSpanInuse)
		},
	},
	"MSpanSys": {
		ID:    "MSpanSys",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.MSpanSys)
		},
	},
	"Mallocs": {
		ID:    "Mallocs",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.Mallocs)
		},
	},
	"NextGC": {
		ID:    "NextGC",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.NextGC)
		},
	},
	"NumForcedGC": {
		ID:    "NumForcedGC",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.NumForcedGC)
		},
	},
	"NumGC": {
		ID:    "NumGC",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.NumGC)
		},
	},
	"OtherSys": {
		ID:    "OtherSys",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.OtherSys)
		},
	},
	"PauseTotalNs": {
		ID:    "PauseTotalNs",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.PauseTotalNs)
		},
	},
	"StackInuse": {
		ID:    "StackInuse",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.StackInuse)
		},
	},
	"StackSys": {
		ID:    "StackSys",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.StackSys)
		},
	},
	"Sys": {
		ID:    "Sys",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.Sys)
		},
	},
	"TotalAlloc": {
		ID:    "TotalAlloc",
		MType: storage.GaugeMetric,
		ExtractorFn: func(stats *runtime.MemStats) string {
			return fmt.Sprintf("%d", stats.TotalAlloc)
		},
	},
}

func getMemMetrics() []storage.Metrics {
	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)
	res := make([]storage.Metrics, 0, len(presets))
	for _, preset := range presets {
		m, err := storage.ParseMetric(preset.MType, preset.ID, preset.ExtractorFn(ms))
		if err != nil {
			continue
		}
		if err = storage.ValidateMetric(m); err == nil {
			res = append(res, *m)
		}
	}
	return res
}

// CollectMemMetrics Every invocation adds metrics to the storage of metrics
func CollectMemMetrics(repository storage.IMetricRepository) {
	for _, metric := range getMemMetrics() {
		if _, err := repository.Collect(&metric); err != nil {
			log.Printf("Problems with saving metric %v", metric)
		}
	}
}

type MetricSender struct {
	ServerAddress string
	*resty.Client
}

func (s *MetricSender) SendMetric(metrics []storage.Metrics) error {
	updateEndpoint := fmt.Sprintf(`%s/update/`, s.ServerAddress)

	bd := storage.MetricsRequest{Metrics: metrics}
	jsonBody, err := json.Marshal(bd)
	if err != nil {
		return err
	}
	var gzipBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&gzipBuffer)
	if _, err = gzipWriter.Write(jsonBody); err != nil {
		return err
	}
	if err = gzipWriter.Close(); err != nil {
		return err
	}
	_, err = s.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(gzipBuffer.Bytes()).
		Post(updateEndpoint)

	return err
}

func formatServerAddress(rawAddress string) string {
	addressWithLocalhost := rawAddress
	if strings.HasPrefix(rawAddress, `:`) {
		addressWithLocalhost = `localhost` + rawAddress
	}

	formattedAddress := addressWithLocalhost
	if !strings.HasPrefix(addressWithLocalhost, `http://`) {
		formattedAddress = `http://` + addressWithLocalhost
	}

	return formattedAddress
}

func NewMetricSender(serverAddress string) *MetricSender {
	formattedServerAddress := formatServerAddress(serverAddress)
	c := resty.New()
	c.SetRetryCount(retries)
	return &MetricSender{
		ServerAddress: formattedServerAddress,
		Client:        c,
	}
}

func SendMetrics(repository storage.IMetricRepository, sender *MetricSender) {
	if err := sender.SendMetric(repository.GetAllMetrics()); err != nil {
		log.Printf("Problems with sending: %v", err)
	}

}
