package server

import (
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string) (int, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	require.NoError(t, err)

	return resp.StatusCode, string(respBody)
}

func TestValueHandler(t *testing.T) {
	repo := storage.NewInMemMetricRepository()
	err := repo.Collect(storage.Counter{Name: "clicks", Value: 5})
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
			statusCode, responseText := testRequest(t, ts, "GET", tc.endpoint)
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
			statusCode, _ := testRequest(t, ts, "POST", tc.endpoint)
			assert.Equal(t, tc.want.code, statusCode)
		})
	}
}
