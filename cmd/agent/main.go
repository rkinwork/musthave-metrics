package main

import (
	"flag"
	"github.com/rkinwork/musthave-metrics/internal/agent"
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"time"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	address := flag.String("a", "", `server host and port`)
	pollInterval := flag.Int("p", 0, `poll interval`)
	reportInterval := flag.Int("r", 0, "time to report in seconds")
	flag.Parse()

	config := New(
		WithAddress(*address),
		WithPollInterval(*pollInterval),
		WithReportInterval(*reportInterval),
	)

	mStorage := storage.GetLocalStorageModel()
	knownMetrics := agent.GetCollectdMetricStorage()
	mSender := agent.MetricSender{ServerAddress: config.address}
	var i = 1
	for {
		if i%int(config.pollInterval/time.Second) == 0 {
			agent.CollectMemMetrics(mStorage, knownMetrics)
		}
		if i%int(config.pollInterval/time.Second) == 0 {
			agent.SendMetrics(mStorage, knownMetrics, mSender)
			i = 0
		}
		time.Sleep(time.Second * 1) // very naive proven
		i += 1
	}
}
