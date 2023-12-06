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
	storage := storage.GetLocalStorageModel()
	updateHandler := server.GetUpdateHandler(storage)
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, updateHandler)
	return http.ListenAndServe(`:8080`, mux)
}
