package server

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rkinwork/musthave-metrics/internal/gzipper"
	"github.com/rkinwork/musthave-metrics/internal/logger"
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"go.uber.org/zap"
	"html/template"
	"log"
	"net/http"
)

const (
	badRequestError         = "mailformed request"
	problemsWithServerError = "problems with server error"
	metricNotFountError     = "metric not found"
)

var indexTemplate = template.Must(template.New("index").Parse(GenerateHTML()))

func NewMetricsRouter(repository storage.IMetricRepository) chi.Router {
	router := chi.NewRouter()
	router.Use(logger.WithLogging)
	router.Use(middleware.Compress(5))
	router.Use(gzipper.CompressedBodyReaderMiddleware)
	router.Get("/", getMainHandler(repository))
	router.Get("/ping", getPingHandler(repository))
	router.Route("/update", func(router chi.Router) {
		router.Post("/", getJSONUpdateHandler(repository))
		router.Post("/{metricType}/{name}/{value}", getUpdateHandler(repository))
	})
	router.Post("/updates/", getJSONUpdateHandler(repository))
	router.Route("/value", func(router chi.Router) {
		router.Post("/", getJSONValueHandler(repository))
		router.Get("/{metricType}/{name}", getValueHandler(repository))
	})
	return router
}

func getMainHandler(repository storage.IMetricRepository) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		metrics := repository.GetAllMetrics()
		writer.Header().Set("Content-Type", "text/html")
		writer.WriteHeader(http.StatusOK)
		if len(metrics) > 0 {
			err := indexTemplate.Execute(writer, metrics)
			logError(0, err)
			return
		}
		logError(writer.Write([]byte("Empty storage")))
	}
}

func getPingHandler(repository storage.IMetricRepository) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html")
		if err := repository.Ping(request.Context()); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}
}

func getValueHandler(repository storage.IMetricRepository) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		metricType, name, value := chi.URLParam(request, "metricType"), chi.URLParam(request, "name"), chi.URLParam(request, "value")
		m, err := storage.ParseMetric(metricType, name, value)
		if err != nil {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		metric, ok := repository.Get(m)
		if !ok {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		writer.Header().Set("Content-Type", "text/plain")
		writer.WriteHeader(http.StatusOK)
		switch m.MType {
		case storage.CounterMetric:
			logError(fmt.Fprintf(writer, "%d", *metric.Delta))
			return
		case storage.GaugeMetric:
			logError(fmt.Fprintf(writer, "%g", *metric.Value))
		}

	}
}

func getJSONValueHandler(repository storage.IMetricRepository) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		contentType := request.Header.Get("Content-type")
		writer.Header().Set("Content-Type", "application/json")
		var errorResp storage.ErrorResponse
		var statusCode = http.StatusOK
		resp := storage.MetricsResponse{ErrorResponse: &errorResp}
		enc := json.NewEncoder(writer)

		defer func() {
			err := request.Body.Close()
			logError(0, err)
		}()
		defer func() {
			writer.WriteHeader(statusCode)
			if err := enc.Encode(resp); err != nil {
				logger.Log.Debug("error encoding response", zap.Error(err))
				return
			}

		}()

		if contentType != "application/json" {
			statusCode = http.StatusUnsupportedMediaType
			errorResp = storage.ErrorResponse{ErrorValue: "unsupported media type"}
			return
		}

		mRequest, err := storage.ParseJSONRequest(request.Body)
		if err != nil {
			statusCode = http.StatusBadRequest
			errorResp = storage.ErrorResponse{ErrorValue: badRequestError}
			return
		}
		metrics, ok := repository.Get(&mRequest.Metrics[0])
		if !ok {
			statusCode = http.StatusNotFound
			errorResp = storage.ErrorResponse{ErrorValue: metricNotFountError}
			return
		}
		resp.Metrics = &metrics
		resp.ErrorResponse = nil

	}
}

func getUpdateHandler(repository storage.IMetricRepository) http.HandlerFunc {
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
		if err = storage.ValidateMetric(metric); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		if _, err = repository.Collect(metric); err == nil {
			writer.WriteHeader(http.StatusOK)
			return
		}
		writer.WriteHeader(http.StatusBadRequest)
	}
}

func getJSONUpdateHandler(repository storage.IMetricRepository) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		contentType := request.Header.Get("Content-type")
		writer.Header().Set("Content-Type", "application/json")
		var errorResp storage.ErrorResponse
		resp := storage.MetricsResponse{ErrorResponse: &errorResp}
		enc := json.NewEncoder(writer)
		var statusCode = http.StatusOK

		defer func() {
			err := request.Body.Close()
			logError(0, err)
		}()
		defer func() {
			writer.WriteHeader(statusCode)
			if err := enc.Encode(resp); err != nil {
				logger.Log.Debug("error encoding response", zap.Error(err))
				return
			}

		}()
		if contentType != "application/json" {
			statusCode = http.StatusUnsupportedMediaType
			errorResp = storage.ErrorResponse{ErrorValue: "unsupported media type"}
			return
		}
		mRequest, err := storage.ParseJSONRequest(request.Body)
		if err != nil {
			statusCode = http.StatusBadRequest
			errorResp = storage.ErrorResponse{ErrorValue: badRequestError}
			return
		}

		var resMetrics = make([]*storage.Metrics, len(mRequest.Metrics))
		for num, mr := range mRequest.Metrics {
			if err = storage.ValidateMetric(&mr); err != nil {
				statusCode = http.StatusBadRequest
				errorResp = storage.ErrorResponse{ErrorValue: badRequestError}
				return
			}

			resMetrics[num], err = repository.Collect(&mr)
			if err != nil {
				statusCode = http.StatusInternalServerError
				errorResp = storage.ErrorResponse{ErrorValue: problemsWithServerError}
				return
			}
		}

		if len(mRequest.Metrics) == 1 {
			resp.Metrics = resMetrics[0]
		}
		resp.ErrorResponse = nil

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
