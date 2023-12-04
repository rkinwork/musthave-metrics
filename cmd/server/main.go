package main

import (
	"net/http"
	"regexp"
)

var validNamePattern = regexp.MustCompile(`^[a-zA-Z]\w{0,127}$`)
var mstore MemStorage

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	mstore = InitLocalMemStorage()
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, UpdateMetric)
	return http.ListenAndServe(`:8080`, mux)
}
