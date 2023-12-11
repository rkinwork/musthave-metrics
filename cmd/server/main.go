package main

import (
	"flag"
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
	address := flag.String("a", "", `server host and port`)
	flag.Parse()
	config := New(WithAddress(*address))
	st := storage.GetLocalStorageModel()
	serverRouter := server.GetMetricsRouter(st)
	return http.ListenAndServe(config.address, serverRouter)
}
