package main

import (
	"context"
	"errors"
	"github.com/rkinwork/musthave-metrics/internal/config"
	"github.com/rkinwork/musthave-metrics/internal/logger"
	"github.com/rkinwork/musthave-metrics/internal/server"
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os/signal"
	"syscall"
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	repository := storage.NewRepository()
	var saver storage.IMetricSaver

	saver, err = storage.NewPgSaver(cnf, repository)
	if err != nil {
		saver = storage.NewFileSaver(*cnf, repository)
	}

	metricSaver := storage.NewMetricsSaver(
		cnf,
		saver,
	)
	metricSaver.Start(ctx)
	serverRouter := server.NewMetricsRouter(metricSaver)
	srv := &http.Server{Addr: cnf.Address, Handler: serverRouter}

	go func() {
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen and serve returned err: %v", err)
		}
	}()
	<-ctx.Done()

	if err = srv.Shutdown(context.TODO()); err != nil { // Use here context with a required timeout
		log.Printf("server shutdown returned an err: %v\n", err)
	}
	err = metricSaver.Done(context.TODO())
	return err
}
