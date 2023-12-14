package storage

import (
	"errors"
	"regexp"
	"strconv"
)

const metricValueMaxLength = 20

var validNamePattern = regexp.MustCompile(`^[a-zA-Z]\w{0,127}$`)
var validGaugePattern = regexp.MustCompile(`^-?\d+(\.\d+)*$`)
var validCounterPattern = regexp.MustCompile(`^\d+$`)

func ParseMetric(valType, name, val string) (Metric, error) {
	var err error
	if !validNamePattern.MatchString(name) {
		return nil, errors.New("not valid metric Name")
	}
	if len(val) > metricValueMaxLength {
		return nil, errors.New("not valid metric Value")
	}
	switch valType {
	case GaugeMetric:
		if !validGaugePattern.MatchString(val) {
			err = errors.New("not valid metric Value")
		}
		if s, err := strconv.ParseFloat(val, 64); err == nil {
			return Gauge{Name: name, Value: s}, nil
		}

	case CounterMetric:
		if !validCounterPattern.MatchString(val) {
			err = errors.New("not valid metric Value")
		}
		if s, err := strconv.ParseInt(val, 10, 64); err == nil {
			return Counter{Name: name, Value: s}, nil
		}
	default:
		err = errors.New("not valid metric type")
	}

	return nil, err
}
