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
	metricType   string
	name         string
	value        string
	defaultValue string
}

// todo Add String()

type MemStorageModelInt interface {
	Add(valType, name, val string) error
	Set(valType, name, val string) error
	Get(valType, name string) (string, bool)
	InsertBy(valType, name, val string) error
	IterMetrics() []MemMetric
}

type MemStorage interface {
	Insert(valType, name, val string) error
	GetOrDefault(valType, name, defVal string) (string, bool)
	GetTypes() []string
	GetNames(metricType string) []string
}

type LocalMemStorage struct {
	m map[string]map[string]string
}

func (s *LocalMemStorage) Insert(valType, name, val string) error {
	s.m[valType][name] = val
	return nil
}

func (s *LocalMemStorage) GetOrDefault(valType, name, defVal string) (string, bool) {
	val, ok := s.m[valType][name]
	if !ok {
		return defVal, ok
	}
	return val, ok
}

func (s *LocalMemStorage) GetTypes() []string {
	metricTypes := make([]string, len(s.m))

	i := 0
	for k := range s.m {
		metricTypes[i] = k
		i++
	}
	return metricTypes
}

func (s *LocalMemStorage) GetNames(metricType string) []string {
	metricNames := make([]string, len(s.m[metricType]))

	i := 0
	for k := range s.m[metricType] {
		metricNames[i] = k
		i++
	}
	return metricNames
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

func sumInt(a, b string) (res string, err error) {
	ap, err := strconv.ParseInt(a, 10, 64)
	if err != nil {
		return res, err
	}
	bp, err := strconv.ParseInt(b, 10, 64)
	if err != nil {
		return res, err
	}
	return fmt.Sprintf("%d", ap+bp), err
}

type MemStorageModel struct {
	storage MemStorage
}

func (m *MemStorageModel) Add(valType, name, val string) error {

	if err := validateMetric(valType, name, val); err != nil {
		return err
	}
	currentVal, _ := m.storage.GetOrDefault(valType, name, defaultMetricVal)
	val, err := sumInt(val, currentVal)
	if err != nil {
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

func (m *MemStorageModel) Get(valType, name string) (string, bool) {
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

func (m *MemStorageModel) IterMetrics() []MemMetric {
	allMetrics := make([]MemMetric, 0)
	for _, metricType := range m.storage.GetTypes() {
		for _, metricName := range m.storage.GetNames(metricType) {
			value, _ := m.storage.GetOrDefault(metricType, metricName, "0")
			allMetrics = append(allMetrics,
				MemMetric{
					metricType: metricType,
					name:       metricName,
					value:      value,
				})
		}
	}
	return allMetrics
}

func GetLocalStorageModel() MemStorageModelInt {
	mMap := map[string]map[string]string{}
	mMap[GaugeMetric] = map[string]string{}
	mMap[CounterMetric] = map[string]string{}

	return &MemStorageModel{storage: &LocalMemStorage{m: mMap}}

}
