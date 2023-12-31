package main

import (
	"github.com/rkinwork/musthave-metrics/internal/agent"
	"github.com/rkinwork/musthave-metrics/internal/config"
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"log"
	"time"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	cnf, err := config.New()
	if err != nil {
		log.Fatalf("problems with config parsing %e", err)
	}

	repository := storage.NewInMemMetricRepository()
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
