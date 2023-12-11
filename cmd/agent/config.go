package main

import (
	"fmt"
	"strings"
	"time"
)

const defaultAddress = `http://localhost:8080`
const defaultReportInterval = 10
const defaultPollInterval = 2

type Config struct {
	address        string
	reportInterval time.Duration
	pollInterval   time.Duration
}

type Option func(config *Config)

func New(options ...Option) Config {
	s := Config{
		address:        defaultAddress,
		reportInterval: defaultReportInterval * time.Second,
		pollInterval:   defaultPollInterval * time.Second,
	}
	for _, fn := range options {
		fn(&s)
	}
	return s
}

func WithAddress(address string) Option {
	return func(s *Config) {
		if address == "" {
			return
		}
		if !strings.HasPrefix(address, `http`) {
			address = `http://` + address
		}
		s.address = address
	}
}

func WithPollInterval(second int) Option {
	return func(s *Config) {
		if second < 0 {
			panic(fmt.Sprintf("poll interval couldn't be negative: %d", second))
		}
		if second > 0 {
			s.pollInterval = time.Duration(second) * time.Second
		}
	}
}

func WithReportInterval(second int) Option {
	return func(s *Config) {
		if second < 0 {
			panic(fmt.Sprintf("report interval couldn't be negative: %d", second))
		}
		if second > 0 {
			s.reportInterval = time.Duration(second) * time.Second
		}
	}
}
