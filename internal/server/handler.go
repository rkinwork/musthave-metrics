package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/rkinwork/musthave-metrics/internal/logger"
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"html/template"
	"log"
	"net/http"
)

var indexTemplate = template.Must(template.New("index").Parse(GenerateHTML()))

func NewMetricsRouter(repository *storage.MetricRepository) chi.Router {
	router := chi.NewRouter()
	router.Use(logger.WithLogging)
	router.Get("/", getMainHandler(repository))
	router.Route("/update", func(router chi.Router) {
		router.Post("/{metricType}/{name}/{value}", getUpdateHandler(repository))
	})
	router.Route("/value", func(router chi.Router) {
		router.Get("/{metricType}/{name}", getValueHandler(repository))
	})
	return router
}

func getMainHandler(repository *storage.MetricRepository) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		metrics := repository.GetAllMetrics()
		writer.WriteHeader(http.StatusOK)
		if len(metrics) > 0 {
			err := indexTemplate.Execute(writer, metrics)
			logError(0, err)
			return
		}
		logError(writer.Write([]byte("Empty storage")))
	}
}

func getValueHandler(repository *storage.MetricRepository) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		metricType, name := chi.URLParam(request, "metricType"), chi.URLParam(request, "name")
		_, err := storage.ParseMetric(metricType, name, "0")
		if err != nil {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		value, ok, err := repository.Get(metricType, name)
		if !ok || err != nil {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		writer.Header().Set("Content-Type", "text/plain")
		writer.WriteHeader(http.StatusOK)
		logError(fmt.Fprintf(writer, "%s", value.ExportValue()))
	}
}

func getUpdateHandler(repository *storage.MetricRepository) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		metricType, name, value := chi.URLParam(request, "metricType"), chi.URLParam(request, "name"), chi.URLParam(request, "value")
		if value == "" {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		metric, err := storage.ParseMetric(metricType, name, value)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		if repository.Collect(metric) == nil {
			writer.WriteHeader(http.StatusOK)
			return
		}
		writer.WriteHeader(http.StatusBadRequest)
	}
}

func logError(_ int, err error) {
	if err != nil {
		log.Printf("An error occurred: %v\n", err)
	}
}

func GenerateHTML() string {
	return `
<b>All Storage Metrics</b>
{{range  .}}
   <li>{{ . }}</li>
{{end}}`
}
