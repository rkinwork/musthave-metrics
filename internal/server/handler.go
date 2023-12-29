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

func NewMetricsRouter(repository *storage.MetricRepository) chi.Router {
	router := chi.NewRouter()
	router.Use(logger.WithLogging)
	router.Use(middleware.Compress(5))
	router.Use(gzipper.CompressedBodyReaderMiddleware)
	router.Get("/", getMainHandler(repository))
	router.Route("/update", func(router chi.Router) {
		router.Post("/", getJSONUpdateHandler(repository))
		router.Post("/{metricType}/{name}/{value}", getUpdateHandler(repository))
	})
	router.Route("/value", func(router chi.Router) {
		router.Post("/", getJSONValueHandler(repository))
		router.Get("/{metricType}/{name}", getValueHandler(repository))
	})
	return router
}

func getMainHandler(repository *storage.MetricRepository) http.HandlerFunc {
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

func getValueHandler(repository *storage.MetricRepository) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		metricType, name := chi.URLParam(request, "metricType"), chi.URLParam(request, "name")
		m, err := storage.ParseMetric(metricType, name, "0")
		if err != nil {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		value, ok, err := repository.Get(m.ExportTypeName(), m.GetName())
		if !ok || err != nil {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		writer.Header().Set("Content-Type", "text/plain")
		writer.WriteHeader(http.StatusOK)
		logError(fmt.Fprintf(writer, "%s", value.ExportValue()))
	}
}

func getJSONValueHandler(repository *storage.MetricRepository) http.HandlerFunc {
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
		value, ok, err := repository.Get(mRequest.MType, mRequest.ID)
		if !ok || err != nil {
			statusCode = http.StatusNotFound
			errorResp = storage.ErrorResponse{ErrorValue: metricNotFountError}
			return
		}
		metrics, err := storage.ConvertToSend(value)
		if err != nil {
			statusCode = http.StatusInternalServerError
			errorResp = storage.ErrorResponse{ErrorValue: problemsWithServerError}

		}
		resp.Metrics = metrics
		resp.ErrorResponse = nil

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
		if _, err = repository.Collect(metric); err == nil {
			writer.WriteHeader(http.StatusOK)
			return
		}
		writer.WriteHeader(http.StatusBadRequest)
	}
}

func getJSONUpdateHandler(repository *storage.MetricRepository) http.HandlerFunc {
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
		m, err := storage.ConvertFrom(mRequest.Metrics)
		if err != nil {
			statusCode = http.StatusBadRequest
			errorResp = storage.ErrorResponse{ErrorValue: badRequestError}
			return
		}
		m, err = repository.Collect(m)
		if err != nil {
			statusCode = http.StatusInternalServerError
			errorResp = storage.ErrorResponse{ErrorValue: problemsWithServerError}
			return
		}
		metrics, err := storage.ConvertToSend(m)
		if err != nil {
			statusCode = http.StatusBadRequest
			errorResp = storage.ErrorResponse{ErrorValue: problemsWithServerError}
		}
		resp.Metrics = metrics
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
