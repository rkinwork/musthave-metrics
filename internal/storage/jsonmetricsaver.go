package storage

import (
	"context"
	"encoding/json"
	"github.com/rkinwork/musthave-metrics/internal/config"
	"github.com/rkinwork/musthave-metrics/internal/logger"
	"go.uber.org/zap"
	"os"
)

type JSONFileSaver struct {
	FilePath string
	IMetricRepository
}

func (js *JSONFileSaver) Save(_ context.Context) error {
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

func (js *JSONFileSaver) Load(_ context.Context) error {
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

func (js *JSONFileSaver) Close(_ context.Context) error {
	return nil
}

func NewFileSaver(cnf config.Config, repository IMetricRepository) *JSONFileSaver {
	return &JSONFileSaver{FilePath: cnf.FileStoragePath, IMetricRepository: repository}
}
