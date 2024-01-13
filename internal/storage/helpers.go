package storage

import (
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strconv"
)

type MetricsRequest struct {
	*Metrics
}

type MetricsResponse struct {
	*Metrics
	*ErrorResponse
}

type ErrorResponse struct {
	ErrorValue string `json:"error"`
}

var validNamePattern = regexp.MustCompile(`^[a-zA-Z]\w{0,127}$`)

func ValidateMetric(m *Metrics) error {
	if !validNamePattern.MatchString(m.ID) {
		return errors.New("not valid metric Name")
	}
	if m.MType == CounterMetric && m.Delta == nil {
		return errors.New("delta value is required for counter metric")
	}
	if m.MType == CounterMetric && *m.Delta < 1 {
		return errors.New("delta value should be positive")
	}
	if m.MType == GaugeMetric && m.Value == nil {
		return errors.New("value is required for gauge metric")
	}
	return nil
}

func ParseMetric(valType, name, val string) (*Metrics, error) {
	res := &Metrics{}
	res.ID = name
	switch valType {
	case GaugeMetric, CounterMetric:
		res.MType = valType
		if val == "" {
			return res, nil
		}
	default:
		return nil, errors.New("not valid metric type")
	}
	var delta *int64
	var value *float64

	if s, err := strconv.ParseFloat(val, 64); err == nil {
		value = &s
	}
	if s, err := strconv.ParseInt(val, 10, 64); err == nil {
		delta = &s
	}
	switch valType {
	case GaugeMetric:
		res.Value = value
	case CounterMetric:
		res.Delta = delta
	}
	return res, nil

}

func ParseJSONRequest(reader io.Reader) (*MetricsRequest, error) {
	var m MetricsRequest
	err := json.NewDecoder(reader).Decode(&m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
