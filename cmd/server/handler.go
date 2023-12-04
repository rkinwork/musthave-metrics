package main

import (
	"net/http"
	"strconv"
	"strings"
)

func UpdateMetric(w http.ResponseWriter, r *http.Request) {
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
	if mValue == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if !validNamePattern.MatchString(mName) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if mType == `counter` {
		v, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		mstore.counter.Add(mName, v)
		w.WriteHeader(http.StatusOK)
		return
	}
	if mType == `gauge` {
		v, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		mstore.gauge.Set(mName, v)
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}
