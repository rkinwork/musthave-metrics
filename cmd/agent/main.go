package main

import (
	"github.com/rkinwork/musthave-metrics/internal/agent"
	"github.com/rkinwork/musthave-metrics/internal/config"
	"github.com/rkinwork/musthave-metrics/internal/logger"
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"go.uber.org/zap"
	"log"
	"time"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	if err := logger.Initialize(zap.InfoLevel.String()); err != nil {
		log.Fatalf("problems with initializing logger %e", err)
	}
	cnf, err := config.NewAgent(true)
	if err != nil {
		log.Fatalf("problems with config parsing %e", err)
	}

	repository := storage.NewRepository()
	sender := agent.NewMetricSender(cnf.Address)
	var i = 1
	for {
		if i%int(cnf.PollInterval/time.Second) == 0 {
			agent.CollectMemMetrics(repository)
		}
		if i%int(cnf.ReportInterval/time.Second) == 0 {
			agent.SendMetrics(repository, sender)
			i = 0
		}
		time.Sleep(time.Second * 1) // very naive proven to errors
		i += 1
	}
}
