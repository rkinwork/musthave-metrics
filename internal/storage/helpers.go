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

func ParseMetric(valType, name, val string) (Metric, error) {
	var delta *int64
	var value *float64
	if s, err := strconv.ParseFloat(val, 64); err == nil {
		value = &s
	}
	if s, err := strconv.ParseInt(val, 10, 64); err == nil {
		delta = &s
	}

	m := NewMetrics(name, valType, delta, value)
	return ConvertFrom(m)

}

func ConvertFrom(m *Metrics) (Metric, error) {
	if !validNamePattern.MatchString(m.ID) {
		return nil, errors.New("not valid metric Name")
	}
	switch m.MType {
	case GaugeMetric:
		if m.Value == nil {
			return nil, errors.New("not valid metric Value")
		}
		return Gauge{Name: m.ID, Value: *m.Value}, nil
	case CounterMetric:
		if m.Delta == nil {
			return nil, errors.New("not valid metric Value")
		}
		if *m.Delta < 0 {
			return nil, errors.New("delta could be only positive")
		}
		return Counter{Name: m.ID, Value: *m.Delta}, nil
	default:
		return nil, errors.New("not valid metric type")
	}
}

func ConvertTo(m Metric) (*Metrics, error) {
	var value *float64
	switch metric := m.(type) {
	case Gauge:
		value = &metric.Value
	case Counter:
		v := float64(metric.Value)
		value = &v
	default:
		return nil, errors.New("unknown metric type")
	}
	return NewMetrics(m.GetName(), m.ExportTypeName(), nil, value), nil
}

func ConvertToSend(m Metric) (*Metrics, error) {
	switch metric := m.(type) {
	case Gauge:
		return NewMetrics(m.GetName(), m.ExportTypeName(), nil, &metric.Value), nil
	case Counter:
		return NewMetrics(m.GetName(), m.ExportTypeName(), &metric.Value, nil), nil
	default:
		return nil, errors.New("unknown metric type")
	}

}

func ParseJSONRequest(reader io.Reader) (*MetricsRequest, error) {
	var m MetricsRequest
	err := json.NewDecoder(reader).Decode(&m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
