package main

import (
	"github.com/rkinwork/musthave-metrics/internal/config"
	"github.com/rkinwork/musthave-metrics/internal/server"
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"log"
	"net/http"
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
	st := storage.NewInMemMetricRepository()
	serverRouter := server.NewMetricsRouter(st)
	return http.ListenAndServe(cnf.Address, serverRouter)
}
