package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"html/template"
	"log"
	"net/http"
)

var html = `
<b>All Storage Metrics</b>
{{range  .}}
   <li>{{ . }}</li>
{{end}}`
var indexTemplate = template.Must(template.New("index").Parse(html))

func NewMetricsRouter(repository *storage.MetricRepository) chi.Router {
	r := chi.NewRouter()
	r.Get("/", getMainHandler(repository))
	r.Route("/update", func(r chi.Router) {
		r.Post("/{metricType}/{name}/{value}", getUpdateHandler(repository))
	})
	r.Route("/value", func(r chi.Router) {
		r.Get("/{metricType}/{name}", getValueHandler(repository))
	})
	return r
}

func getMainHandler(repository *storage.MetricRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var err error
		metrics := repository.IterMetrics()
		w.WriteHeader(http.StatusOK)
		if len(metrics) > 0 {
			err = indexTemplate.Execute(w, metrics)
			return
		}
		_, err = w.Write([]byte("Empty storage"))

		if err != nil {
			log.Printf("problems with hadeling %e", err)
		}

	}
}

func getValueHandler(repository *storage.MetricRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		name := chi.URLParam(r, "name")
		metric, err := storage.ParseMetric(metricType, name, "0")
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		value, ok, err := repository.Get(metric)

		if !ok || err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, err = fmt.Fprintf(w, "%s", value.ExportValue())
		if err != nil {
			log.Printf("problems with writing response %e", err)
		}
	}
}

func getUpdateHandler(repository *storage.MetricRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		metricType := chi.URLParam(r, "metricType")
		name := chi.URLParam(r, "name")
		value := chi.URLParam(r, "value")

		if value == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		metric, err := storage.ParseMetric(metricType, name, value)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := repository.Collect(metric); err == nil {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}
}
