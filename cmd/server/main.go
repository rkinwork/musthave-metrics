package main

import (
	"github.com/rkinwork/musthave-metrics/internal/config"
	"github.com/rkinwork/musthave-metrics/internal/logger"
	"github.com/rkinwork/musthave-metrics/internal/server"
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"go.uber.org/zap"
	"log"
	"net/http"
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
	cnf, err := config.New(true)
	if err != nil {
		log.Fatalf("problems with config parsing %e", err)
	}
	st := storage.NewRepository(cnf)
	serverRouter := server.NewMetricsRouter(st)
	return http.ListenAndServe(cnf.Address, serverRouter)
}
