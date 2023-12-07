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
	updateHandler := server.GetUpdateHandler(storage.GetLocalStorageModel())
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, updateHandler)
	return http.ListenAndServe(`:8080`, mux)
}
