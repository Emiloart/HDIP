package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const serviceName = "verifier-api"

type Config struct {
	ServiceName                   string
	Host                          string
	Port                          int
	LogLevel                      string
	RequestTimeout                time.Duration
	ReadHeaderTimeout             time.Duration
	ShutdownTimeout               time.Duration
	Phase1DatabaseDriver          string
	Phase1DatabaseURL             string
	TrustRegistryBaseURL          string
	TrustRuntimeHydraTokenURL     string
	TrustRuntimeHydraClientID     string
	TrustRuntimeHydraClientSecret string
	TrustRuntimeHydraScope        string
	BuildVersion                  string
}

func Load() (Config, error) {
	port, err := getenvInt("HDIP_PORT", 8082)
	if err != nil {
		return Config{}, err
	}

	requestTimeout, err := getenvDuration("HDIP_REQUEST_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}

	readHeaderTimeout, err := getenvDuration("HDIP_READ_HEADER_TIMEOUT", 3*time.Second)
	if err != nil {
		return Config{}, err
	}

	shutdownTimeout, err := getenvDuration("HDIP_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		ServiceName:                   serviceName,
		Host:                          getenv("HDIP_HOST", "127.0.0.1"),
		Port:                          port,
		LogLevel:                      getenv("HDIP_LOG_LEVEL", "INFO"),
		RequestTimeout:                requestTimeout,
		ReadHeaderTimeout:             readHeaderTimeout,
		ShutdownTimeout:               shutdownTimeout,
		Phase1DatabaseDriver:          getenv("HDIP_PHASE1_DATABASE_DRIVER", "pgx"),
		Phase1DatabaseURL:             getenv("HDIP_PHASE1_DATABASE_URL", ""),
		TrustRegistryBaseURL:          getenv("HDIP_TRUST_REGISTRY_BASE_URL", "http://127.0.0.1:8083"),
		TrustRuntimeHydraTokenURL:     getenv("HDIP_TRUST_RUNTIME_HYDRA_TOKEN_URL", ""),
		TrustRuntimeHydraClientID:     getenv("HDIP_TRUST_RUNTIME_HYDRA_CLIENT_ID", ""),
		TrustRuntimeHydraClientSecret: getenv("HDIP_TRUST_RUNTIME_HYDRA_CLIENT_SECRET", ""),
		TrustRuntimeHydraScope:        getenv("HDIP_TRUST_RUNTIME_HYDRA_SCOPE", "trust.runtime.read"),
		BuildVersion:                  getenv("HDIP_BUILD_VERSION", "dev"),
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
	case strings.TrimSpace(c.TrustRegistryBaseURL) == "":
		return errors.New("trust registry base url must not be empty")
	case strings.TrimSpace(c.TrustRuntimeHydraTokenURL) == "":
		return errors.New("trust runtime hydra token url must be configured")
	case strings.TrimSpace(c.TrustRuntimeHydraClientID) == "":
		return errors.New("trust runtime hydra client id must be configured")
	case strings.TrimSpace(c.TrustRuntimeHydraClientSecret) == "":
		return errors.New("trust runtime hydra client secret must be configured")
	case strings.TrimSpace(c.TrustRuntimeHydraScope) == "":
		return errors.New("trust runtime hydra scope must be configured")
	case strings.TrimSpace(c.Phase1DatabaseURL) == "":
		return errors.New("phase1 sql-primary runtime requires HDIP_PHASE1_DATABASE_URL")
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

func getenvInt(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer, got %q", key, value)
	}

	return parsed, nil
}

func getenvDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration, got %q", key, value)
	}

	return parsed, nil
}
