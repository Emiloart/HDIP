package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const serviceName = "trust-registry"

type Config struct {
	ServiceName       string
	Host              string
	Port              int
	LogLevel          string
	RequestTimeout    time.Duration
	ReadHeaderTimeout time.Duration
	ShutdownTimeout   time.Duration
	BuildVersion      string
}

func Load() (Config, error) {
	cfg := Config{
		ServiceName:       serviceName,
		Host:              getenv("HDIP_HOST", "127.0.0.1"),
		Port:              getenvInt("HDIP_PORT", 8083),
		LogLevel:          getenv("HDIP_LOG_LEVEL", "INFO"),
		RequestTimeout:    getenvDuration("HDIP_REQUEST_TIMEOUT", 5*time.Second),
		ReadHeaderTimeout: getenvDuration("HDIP_READ_HEADER_TIMEOUT", 3*time.Second),
		ShutdownTimeout:   getenvDuration("HDIP_SHUTDOWN_TIMEOUT", 10*time.Second),
		BuildVersion:      getenv("HDIP_BUILD_VERSION", "dev"),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	switch {
	case strings.TrimSpace(c.Host) == "":
		return errors.New("host must not be empty")
	case c.Port <= 0 || c.Port > 65535:
		return fmt.Errorf("port must be between 1 and 65535: %d", c.Port)
	case c.RequestTimeout <= 0:
		return errors.New("request timeout must be positive")
	case c.ReadHeaderTimeout <= 0:
		return errors.New("read header timeout must be positive")
	case c.ShutdownTimeout <= 0:
		return errors.New("shutdown timeout must be positive")
	default:
		return nil
	}
}

func (c Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func getenv(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return fallback
}

func getenvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
