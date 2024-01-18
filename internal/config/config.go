package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Address         string
	ReportInterval  time.Duration
	PollInterval    time.Duration
	StorageType     string
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
	DatabaseDSN     string
}

const (
	defaultAddr            = ":8080"
	defaultReportInterval  = 10 // in seconds
	defaultPollInterval    = 2  // in seconds
	DefaultStorageType     = "filestorage"
	defaultStoreInterval   = 300
	defaultFileStoragePath = "/tmp/metrics-db.json"
	defaultRestore         = true
	developingEnv          = "devStorage"
)

func New(production bool) (*Config, error) {
	cfg := &Config{
		Address:         defaultAddr,
		StorageType:     developingEnv,
		StoreInterval:   defaultStoreInterval,
		FileStoragePath: defaultFileStoragePath,
		Restore:         defaultRestore,
	}
	if production {
		if err := loadFromFlagsServer(cfg); err != nil {
			return nil, err
		}
		if err := loadFromEnv(cfg); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

func NewAgent(production bool) (*Config, error) {
	cfg := &Config{
		Address:         defaultAddr,
		ReportInterval:  defaultReportInterval,
		PollInterval:    defaultPollInterval,
		StorageType:     developingEnv,
		StoreInterval:   defaultStoreInterval,
		FileStoragePath: defaultFileStoragePath,
		Restore:         false,
	}
	if production {
		if err := loadFromFlagsAgent(cfg); err != nil {
			return nil, err
		}
		if err := loadFromEnv(cfg); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

func loadFromEnv(cfg *Config) error {
	parsedConfig := struct {
		Addr            string `env:"ADDRESS"`
		ReportInterval  int64  `env:"REPORT_INTERVAL"`
		PollInterval    int64  `env:"POLL_INTERVAL"`
		StorageType     string `env:"STORAGE_TYPE"`
		StoreInterval   int64  `env:"STORE_INTERVAL"`
		FileStoragePath string `env:"FILE_STORAGE_PATH"`
		Restore         bool   `env:"RESTORE"`
		DatabaseDSN     string `env:"DATABASE_DSN"`
	}{}

	if err := env.Parse(&parsedConfig); err != nil {
		return fmt.Errorf("failed to parse environment variables: %w", err)
	}

	if parsedConfig.Addr != "" {
		cfg.Address = parsedConfig.Addr
	}
	if parsedConfig.StorageType != "" {
		cfg.StorageType = parsedConfig.StorageType
	}
	if parsedConfig.ReportInterval < 0 || parsedConfig.PollInterval < 0 || parsedConfig.StoreInterval < 0 {
		log.Println("negative intervals are not allowed. Use defaults")
	}
	if parsedConfig.ReportInterval > 0 {
		cfg.ReportInterval = time.Duration(parsedConfig.ReportInterval) * time.Second
	}
	if parsedConfig.PollInterval > 0 {
		cfg.PollInterval = time.Duration(parsedConfig.PollInterval) * time.Second
	}
	if parsedConfig.DatabaseDSN != "" {
		cfg.DatabaseDSN = parsedConfig.DatabaseDSN
	}

	return nil
}

func loadFromFlagsAgent(cfg *Config) error {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	addr := flagSet.String("a", defaultAddr, "server host and port")
	reportInterval := flagSet.Int64("r", defaultReportInterval, "How ofter agent should send data to server")
	pollInterval := flagSet.Int64("p", defaultPollInterval, "How often agent should extract metrics")
	storageType := flagSet.String("s", developingEnv, "Storage type configuration")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	cfg.Address = *addr
	cfg.ReportInterval = time.Duration(*reportInterval) * time.Second
	cfg.PollInterval = time.Duration(*pollInterval) * time.Second
	cfg.StorageType = *storageType

	return nil
}

func loadFromFlagsServer(cfg *Config) error {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	addr := flagSet.String("a", defaultAddr, "server host and port")

	storageType := flagSet.String("s", DefaultStorageType, "Storage type configuration")
	fileStoragePath := flagSet.String("f", defaultFileStoragePath, "File storage path")
	restore := flagSet.Bool("r", defaultRestore, "Is restore metrics from file storage")
	storeInterval := flagSet.Int64("i", defaultStoreInterval, "How often agent should dump metrics")
	dbdsn := flagSet.String("d", "", "Database DSN in URL format")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	cfg.Address = *addr

	cfg.StorageType = *storageType
	cfg.FileStoragePath = *fileStoragePath
	cfg.Restore = *restore
	cfg.StoreInterval = time.Duration(*storeInterval) * time.Second
	cfg.DatabaseDSN = *dbdsn

	return nil
}
