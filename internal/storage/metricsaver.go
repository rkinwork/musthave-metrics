package storage

import (
	"context"
	"encoding/json"
	"github.com/rkinwork/musthave-metrics/internal/config"
	"github.com/rkinwork/musthave-metrics/internal/logger"
	"go.uber.org/zap"
	"os"
	"time"
)

type ILoadSave interface {
	Save() error
	Load() error
}

type IMetricSaver interface {
	IMetricRepository
	ILoadSave
}

type JSONFileSaver struct {
	FilePath string
	IMetricRepository
}

type NoopMetricSaver struct{}

func (js *JSONFileSaver) Save() error {
	file, err := os.OpenFile(js.FilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			logger.Log.Error("problems with closing file", zap.Error(err))
		}
	}(file)

	bytes, err := json.Marshal(js.GetAllMetrics())
	if err != nil {
		return err
	}

	_, err = file.Write(bytes)
	return err
}

func (js *JSONFileSaver) Load() error {
	var metrics []Metrics
	if _, err := os.Stat(js.FilePath); os.IsNotExist(err) {
		return nil
	}
	file, err := os.Open(js.FilePath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			logger.Log.Error("problems with closing file", zap.Error(err))
		}
	}(file)

	data := json.NewDecoder(file)
	err = data.Decode(&metrics)

	if err != nil {
		return err
	}
	for _, metric := range metrics {
		if _, err := js.Set(&metric); err != nil {
			logger.Log.Error("error while setting metric", zap.Error(err))
		}
	}

	return nil
}

type MetricsSaver struct {
	config *config.Config
	ticker *time.Ticker
	quit   chan struct{}
	IMetricSaver
}

func (ms *MetricsSaver) Done() {
	<-ms.quit
}

func (ms *MetricsSaver) Start(ctx context.Context) {
	if ms.config.Restore {
		_ = ms.Load()
	}
	go func() {
		for {
			select {
			case <-ms.ticker.C:
				_ = ms.Save()
			case <-ctx.Done():
				ms.ticker.Stop()
				_ = ms.Save()
				close(ms.quit)
				return
			}
		}
	}()
}

func (ms *MetricsSaver) Collect(metric *Metrics) (*Metrics, error) {

	m, err := ms.IMetricSaver.Collect(metric)
	if ms.config.StoreInterval == 0 {
		_ = ms.Save()
	}
	return m, err
}

func NewMetricsSaver(config *config.Config, repo IMetricSaver) *MetricsSaver {
	var ticker *time.Ticker
	if config.StoreInterval > 0 {
		ticker = time.NewTicker(config.StoreInterval)
	}

	ms := &MetricsSaver{
		config:       config,
		ticker:       ticker,
		quit:         make(chan struct{}),
		IMetricSaver: repo,
	}
	return ms
}
