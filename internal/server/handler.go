package server

import (
	"errors"
	"github.com/rkinwork/musthave-metrics/internal/storage"
	"net/http"
	"strings"
)

func GetUpdateHandler(repository storage.MemStorageModelInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		if len(r.Form) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("Parameters are not allowed in request"))
			return
		}
		var data [4]string
		params := strings.Split(strings.Trim(r.URL.Path, `/`), `/`)
		for n, el := range params {
			if n >= len(data) {
				break
			}
			data[n] = el
		}
		mType, mName, mValue := data[1], data[2], data[3]

		var err error
		switch {
		case mValue == "":
			w.WriteHeader(http.StatusNotFound)
			return

		case mType == storage.GaugeMetric:
			err = repository.Add(storage.GaugeMetric, mName, mValue)

		case mType == storage.CounterMetric:
			err = repository.Set(storage.CounterMetric, mName, mValue)

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
