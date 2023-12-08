package main

import (
	"github.com/rkinwork/musthave-metrics/internal/server"
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"net/http"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	st := storage.GetLocalStorageModel()
	serverRouter := server.GetMetricsRouter(st)
	return http.ListenAndServe(`:8080`, serverRouter)
}
