package storage

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

const (
	GaugeMetric   = `gauge`
	CounterMetric = `counter`
)

const defaultMetricVal = `0`
const metricValueMaxLength = 20

var validNamePattern = regexp.MustCompile(`^[a-zA-Z]\w{0,127}$`)
var validGaugePattern = regexp.MustCompile(`^-?\d+(\.\d+)*$`)
var validCounterPattern = regexp.MustCompile(`^\d+$`)

type MemMetric struct {
	valType string
	name    string
	value   string
}
type MemStorageModelInt interface {
	Add(valType, name, val string) error
	Set(valType, name, val string) error
	Get(valType, name string) (string, error)
	InsertBy(valType, name, val string) error
}

type MemStorage interface {
	Insert(valType, name, val string) error
	GetOrDefault(valType, name, defVal string) (string, error)
}

type LocalMemStorage struct {
	m map[string]map[string]string
}

func (s *LocalMemStorage) Insert(valType, name, val string) error {
	s.m[valType][name] = val
	return nil
}

func (s *LocalMemStorage) GetOrDefault(valType, name, defVal string) (string, error) {
	val, ok := s.m[valType][name]
	if !ok {
		return defVal, nil
	}
	return val, nil
}

func validateMetric(valType, name, val string) error {
	var err error
	if !validNamePattern.MatchString(name) {
		return errors.New("not valid metric name")
	}
	if len(val) > metricValueMaxLength {
		return errors.New("not valid metric value")
	}
	switch valType {
	case GaugeMetric:
		if !validGaugePattern.MatchString(val) {
			err = errors.New("not valid metric value")
		}
	case CounterMetric:
		if !validCounterPattern.MatchString(val) {
			err = errors.New("not valid metric value")
		}
	default:
		err = errors.New("not valid metric type")
	}
	if err != nil {
		return err
	}
	return err
}

func sumGauge(a, b string) (res string, err error) {
	ap, err := strconv.ParseFloat(a, 64)
	if err != nil {
		return res, err
	}
	bp, err := strconv.ParseFloat(b, 64)
	if err != nil {
		return res, err
	}
	return fmt.Sprintf("%f", ap+bp), err

}

type MemStorageModel struct {
	storage MemStorage
}

func (m *MemStorageModel) Add(valType, name, val string) error {

	if err := validateMetric(valType, name, val); err != nil {
		return err
	}
	oldVal, err := m.storage.GetOrDefault(valType, name, defaultMetricVal)
	if err != nil {
		return err
	}
	if val, err = sumGauge(val, oldVal); err != nil {
		return err
	}
	return m.storage.Insert(valType, name, val)
}

func (m *MemStorageModel) Set(valType, name, val string) error {
	if err := validateMetric(valType, name, val); err != nil {
		return err
	}
	return m.storage.Insert(valType, name, val)
}

func (m *MemStorageModel) Get(valType, name string) (string, error) {
	return m.storage.GetOrDefault(valType, name, defaultMetricVal)
}

func (m *MemStorageModel) InsertBy(valType, name, val string) error {
	if err := validateMetric(valType, name, val); err != nil {
		return err
	}
	switch valType {
	case CounterMetric:
		return m.Add(valType, name, val)
	case GaugeMetric:
		return m.Set(valType, name, val)
	}
	return errors.New(`unknown metric type`)
}

func GetLocalStorageModel() MemStorageModelInt {
	mMap := map[string]map[string]string{}
	mMap[GaugeMetric] = map[string]string{}
	mMap[CounterMetric] = map[string]string{}

	return &MemStorageModel{storage: &LocalMemStorage{m: mMap}}

}
