package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rkinwork/musthave-metrics/internal/config"
	"github.com/rkinwork/musthave-metrics/internal/logger"
	"go.uber.org/zap"
	"os"
	"sync"
	"time"
)

// Storage and saving interfaces and implementations

type MetricSaver interface {
	Save([]Metrics) error
	Load() ([]Metrics, error)
}

type JSONFileSaver struct {
	FilePath string
}

type NoopMetricSaver struct{}

func (js *JSONFileSaver) Save(metrics []Metrics) error {
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

	bytes, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	_, err = file.Write(bytes)
	return err
}

func (js *JSONFileSaver) Load() ([]Metrics, error) {
	file, err := os.Open(js.FilePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			logger.Log.Error("problems with closing file", zap.Error(err))
		}
	}(file)

	var metrics []Metrics
	data := json.NewDecoder(file)
	err = data.Decode(&metrics)

	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (*NoopMetricSaver) Save(_ []Metrics) error { return nil }

func (*NoopMetricSaver) Load() ([]Metrics, error) { return nil, nil }

// In-memory storage

type InMemMetricStorage struct {
	m        map[MetricHash]Metrics
	Interval time.Duration
	Restore  bool
	sync.Mutex
	save  chan bool
	saver MetricSaver
}

func (i *InMemMetricStorage) Get(m Metrics) (Metrics, bool) {
	res, ok := i.m[m.GetHash()]
	return res, ok
}

func (i *InMemMetricStorage) Set(m Metrics) error {
	i.Lock()
	i.m[m.GetHash()] = m
	i.Unlock()
	i.save <- true
	return nil
}

func (i *InMemMetricStorage) Delete(m Metrics) error {
	i.Lock()
	delete(i.m, m.GetHash())
	i.Unlock()
	i.save <- true
	return nil
}

func (i *InMemMetricStorage) IterMetrics() []Metrics {
	var res []Metrics
	for _, metric := range i.m {
		res = append(res, metric)
	}
	return res
}

func (i *InMemMetricStorage) SaveMetrics() error {
	i.Lock()
	err := i.saver.Save(i.IterMetrics())
	i.Unlock()
	return err
}

func (i *InMemMetricStorage) LoadMetrics() error {
	if !i.Restore {
		return nil
	}

	metrics, err := i.saver.Load()
	if err != nil {
		return err
	}

	for _, m := range metrics {
		_ = i.Set(m)
	}

	return nil
}

func NewInMemMetricStorage(ctx context.Context, cfg *config.Config) *InMemMetricStorage {

	var saver MetricSaver
	switch cfg.StorageType {
	case config.DefaultStorageType:
		saver = &JSONFileSaver{FilePath: cfg.FileStoragePath}
	default:
		saver = &NoopMetricSaver{}
	}
	imms := &InMemMetricStorage{
		m:        make(map[MetricHash]Metrics),
		Interval: cfg.StoreInterval,
		Restore:  cfg.Restore,
		save:     make(chan bool, 1),
		saver:    saver,
	}

	go func() {
		ticker := time.NewTicker(imms.Interval)
		for {
			select {
			case <-ticker.C:
				if err := imms.SaveMetrics(); err != nil {
					fmt.Println("error while saving", err)
				}
			case <-imms.save:
				if imms.Interval == 0 {
					if err := imms.SaveMetrics(); err != nil {
						fmt.Println("error while saving", err)
					}
				}
			case <-ctx.Done(): // listen for context cancellation here
				ticker.Stop()
				if err := imms.SaveMetrics(); err != nil {
					fmt.Println("error while saving", err)
				}
				return
			}
		}
	}()

	if err := imms.LoadMetrics(); err != nil {
		fmt.Println("error while saving", err)
	}

	return imms
}
