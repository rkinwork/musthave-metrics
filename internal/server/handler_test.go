package server

import (
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUpdateHandlerHandler(t *testing.T) {
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
				code: 200,
			},
		},
		{
			name:     "positive flow gauge",
			endpoint: "/update/gauge/roman/868434",
			want: want{
				code: 200,
			},
		},
		{
			name:     "positive flow gauge # 2",
			endpoint: "/update/gauge/roman/868434.12",
			want: want{
				code: 200,
			},
		},
		{
			name:     "wrong counter value type with letters",
			endpoint: "/update/counter/roman/868434sdf",
			want: want{
				code: 400,
			},
		},
		{
			name:     "wrong metric name",
			endpoint: "/update/counter/1roman/868434sdf",
			want: want{
				code: 400,
			},
		},
		{
			name:     "wrong counter value type with letters",
			endpoint: "/update/counter/roman/23423sdf",
			want: want{
				code: 400,
			},
		},
		{
			name:     "wrong counter type",
			endpoint: "/update/counter/roman/868434.2342",
			want: want{
				code: 400,
			},
		},
		{
			name:     "absent metric value",
			endpoint: "/update/counter/roman/",
			want: want{
				code: 404,
			},
		},
		{
			name:     "wrong counter type",
			endpoint: "/update/unknown/roman/868434sdf",
			want: want{
				code: 400,
			},
		},
		{
			name:     "wrong counter type",
			endpoint: "/update/unknown/roman/868434sdf",
			want: want{
				code: 400,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.endpoint, nil)
			w := httptest.NewRecorder()
			handler := GetUpdateHandler(storage.GetLocalStorageModel())
			handler(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			_ = res.Body.Close()
		})
	}
}
