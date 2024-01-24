package storage

import (
	"context"
	"github.com/avast/retry-go/v4"
	"github.com/rkinwork/musthave-metrics/internal/config"
	"log"
	"time"
)

type ILoadSaveClose interface {
	Save(ctx context.Context) error
	Load(ctx context.Context) error
	Close(ctx context.Context) error
}

type IMetricSaver interface {
	IMetricRepository
	ILoadSaveClose
}

type NoopMetricSaver struct{}

type MetricsSaver struct {
	config *config.Config
	ticker *time.Ticker
	quit   chan struct{}
	IMetricSaver
}

func (ms *MetricsSaver) Done(ctx context.Context) error {
	<-ms.quit
	return ms.Close(ctx)
}

func (ms *MetricsSaver) Start(ctx context.Context) error {
	if ms.config.Restore {
		err := retry.Do(func() error { return ms.Load(ctx) })
		if err != nil {
			log.Println(err)
			return err
		}

	}
	go func() {
		for {
			select {
			case <-ms.ticker.C:
				_ = ms.Save(ctx)
			case <-ctx.Done():
				ms.ticker.Stop()
				_ = ms.Save(ctx)
				close(ms.quit)
				return
			}
		}
	}()
	return nil
}

func (ms *MetricsSaver) Collect(metric *Metrics) (*Metrics, error) {

	m, err := ms.IMetricSaver.Collect(metric)
	if ms.config.StoreInterval == 0 {
		_ = ms.Save(context.TODO())
	}
	return m, err
}

func NewMetricsSaver(config *config.Config, repo IMetricSaver) *MetricsSaver {
	var ticker = new(time.Ticker)
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
