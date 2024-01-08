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
	"os"
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
	// Create a new context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a channel to listen for OS signals
	sigs := make(chan os.Signal, 1)

	// Register the channel to receive SIGINT and SIGTERM signals
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	st := storage.NewRepository(ctx, cnf)
	serverRouter := server.NewMetricsRouter(st)
	srv := &http.Server{Addr: cnf.Address, Handler: serverRouter}

	// Start a goroutine to handle the signal
	go func() {
		sig := <-sigs
		logger.Log.Info("Received signal, Exiting...", zap.String("signal", sig.String()))
		srv.Shutdown(ctx)
		cancel() // Invoke cancel on receiving a signal

	}()

	if serr := srv.ListenAndServe(); serr != nil && !errors.Is(serr, http.ErrServerClosed) {
		err = serr
	}
	return err
}
