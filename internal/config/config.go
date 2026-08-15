package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Port         int
	DataDir      string
	DatabasePath string
	ScanWorkers  int
	LogLevel     slog.Level
}

func Load() (Config, error) {
	dataDir := env("TELECOM_DATA_DIR", "./data")
	port, err := envInt("TELECOM_PORT", 14000, 1, 65535)
	if err != nil {
		return Config{}, err
	}
	workers, err := envInt("TELECOM_SCAN_WORKERS", 32, 1, 512)
	if err != nil {
		return Config{}, err
	}
	level, err := parseLevel(env("TELECOM_LOG_LEVEL", "info"))
	if err != nil {
		return Config{}, err
	}
	if err := os.MkdirAll(filepath.Join(dataDir, "attachments"), 0o750); err != nil {
		return Config{}, fmt.Errorf("create data directory: %w", err)
	}
	return Config{Port: port, DataDir: dataDir, DatabasePath: filepath.Join(dataDir, "telecom.sqlite"), ScanWorkers: workers, LogLevel: level}, nil
}

func (c Config) HTTPAddress() string { return fmt.Sprintf(":%d", c.Port) }

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback, minimum, maximum int) (int, error) {
	value, err := strconv.Atoi(env(key, strconv.Itoa(fallback)))
	if err != nil || value < minimum || value > maximum {
		return 0, fmt.Errorf("%s must be an integer between %d and %d", key, minimum, maximum)
	}
	return value, nil
}

func parseLevel(value string) (slog.Level, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(value)); err != nil {
		return 0, fmt.Errorf("TELECOM_LOG_LEVEL must be debug, info, warn, or error")
	}
	return level, nil
}
