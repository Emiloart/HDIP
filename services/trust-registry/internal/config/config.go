package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const serviceName = "trust-registry"

type Config struct {
	ServiceName                       string
	Host                              string
	Port                              int
	LogLevel                          string
	RequestTimeout                    time.Duration
	ReadHeaderTimeout                 time.Duration
	ShutdownTimeout                   time.Duration
	Phase1DatabaseDriver              string
	Phase1DatabaseURL                 string
	Phase1StatePath                   string
	TrustBootstrapPath                string
	TrustRuntimeHydraIntrospectionURL string
	TrustRuntimeHydraClientID         string
	TrustRuntimeHydraClientSecret     string
	TrustRuntimeHydraExpectedClientID string
	TrustRuntimeHydraRequiredScope    string
	BuildVersion                      string
}

func Load() (Config, error) {
	port, err := getenvInt("HDIP_PORT", 8083)
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
		ServiceName:                       serviceName,
		Host:                              getenv("HDIP_HOST", "127.0.0.1"),
		Port:                              port,
		LogLevel:                          getenv("HDIP_LOG_LEVEL", "INFO"),
		RequestTimeout:                    requestTimeout,
		ReadHeaderTimeout:                 readHeaderTimeout,
		ShutdownTimeout:                   shutdownTimeout,
		Phase1DatabaseDriver:              getenv("HDIP_PHASE1_DATABASE_DRIVER", "pgx"),
		Phase1DatabaseURL:                 getenv("HDIP_PHASE1_DATABASE_URL", ""),
		Phase1StatePath:                   phase1StatePath(),
		TrustBootstrapPath:                getenv("HDIP_TRUST_REGISTRY_BOOTSTRAP_PATH", ""),
		TrustRuntimeHydraIntrospectionURL: getenv("HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_URL", ""),
		TrustRuntimeHydraClientID:         getenv("HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_CLIENT_ID", ""),
		TrustRuntimeHydraClientSecret:     getenv("HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_CLIENT_SECRET", ""),
		TrustRuntimeHydraExpectedClientID: getenv("HDIP_TRUST_RUNTIME_HYDRA_EXPECTED_CLIENT_ID", "verifier-api"),
		TrustRuntimeHydraRequiredScope:    getenv("HDIP_TRUST_RUNTIME_HYDRA_REQUIRED_SCOPE", "trust.runtime.read"),
		BuildVersion:                      getenv("HDIP_BUILD_VERSION", "dev"),
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
	case strings.TrimSpace(c.Phase1DatabaseURL) == "" && strings.TrimSpace(c.Phase1StatePath) == "":
		return errors.New("phase1 database url or transitional state path must be configured")
	case strings.TrimSpace(c.TrustRuntimeHydraIntrospectionURL) == "":
		return errors.New("trust runtime hydra introspection url must be configured")
	case strings.TrimSpace(c.TrustRuntimeHydraClientID) == "":
		return errors.New("trust runtime hydra introspection client id must be configured")
	case strings.TrimSpace(c.TrustRuntimeHydraClientSecret) == "":
		return errors.New("trust runtime hydra introspection client secret must be configured")
	case strings.TrimSpace(c.TrustRuntimeHydraExpectedClientID) == "":
		return errors.New("trust runtime hydra expected client id must be configured")
	case strings.TrimSpace(c.TrustRuntimeHydraRequiredScope) == "":
		return errors.New("trust runtime hydra required scope must be configured")
	case strings.TrimSpace(c.Phase1DatabaseURL) != "" && strings.TrimSpace(c.TrustBootstrapPath) != "":
		return errors.New("trust registry bootstrap path must be applied through the phase1sql CLI when the primary sql path is enabled")
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

func phase1StatePath() string {
	if value := strings.TrimSpace(os.Getenv("HDIP_PHASE1_STATE_PATH")); value != "" {
		return value
	}

	return getenv("HDIP_PHASE1_RUNTIME_PATH", defaultPhase1StatePath())
}

func defaultPhase1StatePath() string {
	return filepath.Join(os.TempDir(), "hdip-phase1-state.json")
}
