package server

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"html/template"
	"log"
	"net/http"
)

func GetMetricsRouter(repository storage.MemStorageModelInt) chi.Router {
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

func getMainHandler(repository storage.MemStorageModelInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		html := `<b>hello</b>
{{range  .}}
   <li>{{ . }}</li>
{{end}}`
		tmpl, err := template.New("index").Parse(html)
		if err != nil {
			log.Fatal(err)
			return
		}
		metrics := repository.IterMetrics()
		w.WriteHeader(http.StatusOK)
		if len(metrics) > 0 {
			_ = tmpl.Execute(w, metrics)
			return
		}
		w.Write([]byte("empty response"))

	}
}

func getValueHandler(repository storage.MemStorageModelInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		name := chi.URLParam(r, "name")
		value, ok := repository.Get(metricType, name)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "%s", value)
	}
}

func getUpdateHandler(repository storage.MemStorageModelInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		metricType := chi.URLParam(r, "metricType")
		name := chi.URLParam(r, "name")
		value := chi.URLParam(r, "value")

		var err error
		switch {
		case value == "":
			w.WriteHeader(http.StatusNotFound)
			return

		case metricType == storage.GaugeMetric:
			err = repository.Set(storage.GaugeMetric, name, value)

		case metricType == storage.CounterMetric:
			err = repository.Add(storage.CounterMetric, name, value)

		default:
			err = errors.New(`unknown metric`)
		}

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)

	}
}
