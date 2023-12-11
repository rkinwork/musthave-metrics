package main

import "os"

const defaultAddress = `:8080`

type Config struct {
	address string
}

type Option func(config *Config)

func New(options ...Option) Config {

	s := Config{
		address: defaultAddress,
	}
	for _, fn := range options {
		fn(&s)
	}
	return s
}

func WithAddress(address string) Option {
	return func(s *Config) {
		if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
			address = envAddress
		}
		if address == "" {
			return
		}
		s.address = address
	}
}
