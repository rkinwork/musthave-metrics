package server

import (
	"bytes"
	"encoding/json"
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, header http.Header, body io.Reader) (int, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)
	req.Header = header

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	require.NoError(t, err)

	return resp.StatusCode, string(respBody)
}

func TestValueHandler(t *testing.T) {
	repo := storage.NewInMemMetricRepository()
	_, err := repo.Collect(storage.Counter{Name: "clicks", Value: 5})
	require.NoError(t, err)
	ts := httptest.NewServer(NewMetricsRouter(repo))
	defer ts.Close()
	type want struct {
		code         int
		responseText string
	}
	tests := []struct {
		name     string
		endpoint string
		want     want
	}{
		{
			name:     "positive flow counter",
			endpoint: "/value/counter/clicks",
			want: want{
				code:         http.StatusOK,
				responseText: "5",
			},
		},
		{
			name:     "positive flow counter 2",
			endpoint: "/value/counter/clicks/",
			want: want{
				code:         http.StatusNotFound,
				responseText: "404 page not found\n",
			},
		},
		{
			name:     "negative flow counter",
			endpoint: "/value/counter/unknown",
			want: want{
				code:         http.StatusNotFound,
				responseText: "",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			statusCode, responseText := testRequest(t, ts, "GET", tc.endpoint, http.Header{}, nil)
			assert.Equal(t, tc.want.code, statusCode)
			assert.Equal(t, tc.want.responseText, responseText)

		})
	}
}

func TestUpdateHandler(t *testing.T) {
	repo := storage.NewInMemMetricRepository()
	ts := httptest.NewServer(NewMetricsRouter(repo))
	defer ts.Close()
	type want struct {
		code int
	}
	tests := []struct {
		name     string
		endpoint string
		want     want
	}{
		{
			name:     "positive flow counter",
			endpoint: "/update/counter/roman/868434",
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:     "positive flow gauge",
			endpoint: "/update/gauge/roman/868434",
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:     "positive flow gauge # 2",
			endpoint: "/update/gauge/roman/868434.12",
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:     "wrong counter value type with letters",
			endpoint: "/update/counter/roman/868434sdf",
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:     "wrong metric name",
			endpoint: "/update/counter/1roman/868434sdf",
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:     "wrong counter value type with letters",
			endpoint: "/update/counter/roman/23423sdf",
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:     "wrong counter type",
			endpoint: "/update/counter/roman/868434.2342",
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:     "absent metric value",
			endpoint: "/update/counter/roman/",
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:     "wrong counter type",
			endpoint: "/update/unknown/roman/868434sdf",
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:     "wrong counter type",
			endpoint: "/update/unknown/roman/868434sdf",
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			statusCode, _ := testRequest(t, ts, "POST", tc.endpoint, http.Header{}, nil)
			assert.Equal(t, tc.want.code, statusCode)
		})
	}
}

func TestJSONUpdateHandler(t *testing.T) {
	repo := storage.NewInMemMetricRepository()
	ts := httptest.NewServer(NewMetricsRouter(repo))
	var icnt int64 = 1
	var fcnt float64 = 1

	defer ts.Close()
	type want struct {
		code int
		resp storage.MetricsResponse
	}
	tests := []struct {
		name    string
		payload storage.MetricsRequest
		want    want
	}{
		{
			name: "positive flow counter",
			payload: storage.MetricsRequest{
				Metrics: storage.NewMetrics("testid", storage.CounterMetric, &icnt, nil),
			},
			want: want{
				code: http.StatusOK,
				resp: storage.MetricsResponse{
					Metrics: storage.NewMetrics("testid", storage.CounterMetric, nil, &fcnt),
				},
			},
		},
		{
			name: "positive flow gauge",
			payload: storage.MetricsRequest{
				Metrics: storage.NewMetrics("testid", storage.GaugeMetric, nil, &fcnt),
			},
			want: want{
				code: http.StatusOK,
				resp: storage.MetricsResponse{
					Metrics: storage.NewMetrics("testid", storage.GaugeMetric, nil, &fcnt),
				},
			},
		},
		{
			name: "negative flow gauge",
			payload: storage.MetricsRequest{
				Metrics: storage.NewMetrics("testid", storage.GaugeMetric, &icnt, nil),
			},
			want: want{
				code: http.StatusBadRequest,
				resp: storage.MetricsResponse{
					Metrics:       nil,
					ErrorResponse: &storage.ErrorResponse{ErrorValue: badRequestError},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := bytes.NewBuffer(make([]byte, 0))
			enc := json.NewEncoder(w)
			enc.Encode(tc.payload)
			statusCode, bd := testRequest(t, ts, "POST", "/update/", http.Header{
				"Content-Type": {"application/json"},
			}, w)
			w.Reset()
			enc.Encode(tc.want.resp)
			assert.Equal(t, tc.want.code, statusCode)
			assert.JSONEq(t, w.String(), bd)
		})
	}
}
