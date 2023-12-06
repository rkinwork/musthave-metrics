package main

import (
	"github.com/rkinwork/musthave-metrics/internal/agent"
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"time"
)

const (
	pollInterval   = 2
	reportInterval = 10
	serverAddress  = "http://localhost:8080"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {

	mStorage := storage.GetLocalStorageModel()
	knownMetrics := agent.GetCollectdMetricStorage()
	var i = 1
	for {
		if i%pollInterval == 0 {
			agent.CollectMemMetrics(mStorage, knownMetrics)
		}
		if i%reportInterval == 0 {
			agent.SendMetrics(mStorage, knownMetrics, serverAddress)
			i = 0
		}
		time.Sleep(time.Second * 1) // very naive could un sync
		i += 1
	}
}
