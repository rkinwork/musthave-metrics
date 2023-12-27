package server

import (
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
	_, err := repo.Collect(storage.Counter{Name: "pollcount", Value: 1})
	require.NoError(t, err)

	defer ts.Close()
	type want struct {
		code int
		resp string
	}
	tests := []struct {
		name    string
		payload string
		want    want
	}{
		{
			name:    "positive flow counter",
			payload: `{"id": "test", "type":"counter", "delta": 1}`,
			want: want{
				code: http.StatusOK,
				resp: `{"id": "test", "type":"counter", "delta": 1}`,
			},
		},
		{
			name:    "positive flow gauge",
			payload: `{"id": "test", "type":"gauge", "value": 1}`,
			want: want{
				code: http.StatusOK,
				resp: `{"id": "test", "type":"gauge", "value": 1}`,
			},
		},
		{
			name:    "negative flow gauge",
			payload: `{"id": "test", "type":"gauge", "delta": 1}`,
			want: want{
				code: http.StatusBadRequest,
				resp: `{"error":"mailformed request"}`,
			},
		},
		{
			name:    "negative flow counter float delta",
			payload: `{"id": "test", "type":"counter", "delta": 1.23}`,
			want: want{
				code: http.StatusBadRequest,
				resp: `{"error":"mailformed request"}`,
			},
		},
		{
			name:    "negative flow counter negative delta",
			payload: `{"id": "test", "type":"counter", "delta": -1}`,
			want: want{
				code: http.StatusBadRequest,
				resp: `{"error":"mailformed request"}`,
			},
		},
		{
			name:    "positive poll counter increment",
			payload: `{"id": "pollcount", "type":"counter", "delta": 1}`,
			want: want{
				code: http.StatusOK,
				resp: `{"id": "pollcount", "type":"counter", "delta": 2}`,
			},
		},
		{
			name:    "empty payload",
			payload: ``,
			want: want{
				code: http.StatusBadRequest,
				resp: `{"error":"mailformed request"}`,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := strings.NewReader(tc.payload)
			statusCode, bd := testRequest(t, ts, "POST", "/update/", http.Header{
				"Content-Type": {"application/json"},
			}, w)
			assert.Equal(t, tc.want.code, statusCode)
			assert.JSONEq(t, tc.want.resp, bd)
		})
	}
}

func TestJSONValueHandler(t *testing.T) {
	repo := storage.NewInMemMetricRepository()
	_, err := repo.Collect(storage.Counter{Name: "test", Value: 1})
	require.NoError(t, err)
	_, err = repo.Collect(storage.Gauge{Name: "test", Value: 1})
	require.NoError(t, err)
	ts := httptest.NewServer(NewMetricsRouter(repo))
	defer ts.Close()
	type want struct {
		code int
		resp string
	}
	tests := []struct {
		name    string
		payload string
		want    want
	}{
		{
			name:    "positive flow counter",
			payload: `{"id": "test", "type":"counter"}`,
			want: want{
				code: http.StatusOK,
				resp: `{"id": "test", "type":"counter", "delta": 1}`,
			},
		},
		{
			name:    "positive flow gauge",
			payload: `{"id": "test", "type":"gauge"}`,
			want: want{
				code: http.StatusOK,
				resp: `{"id": "test", "type":"gauge", "value": 1}`,
			},
		},
		{
			name:    "unknown type",
			payload: `{"id": "test", "type":"badtype"}`,
			want: want{
				code: http.StatusNotFound,
				resp: `{"error":"metric not found"}`,
			},
		},
		{
			name:    "absent id",
			payload: `{"id": "unknown", "type":"counter"}`,
			want: want{
				code: http.StatusNotFound,
				resp: `{"error":"metric not found"}`,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := strings.NewReader(tc.payload)
			statusCode, bd := testRequest(t, ts, "POST", "/value/", http.Header{
				"Content-Type": {"application/json"},
			}, w)
			assert.Equal(t, tc.want.code, statusCode)
			assert.JSONEq(t, tc.want.resp, bd)
		})
	}
}
