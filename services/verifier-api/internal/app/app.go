package app

import (
	"context"

	"github.com/Emiloart/HDIP/packages/go/foundation/observability"
	"github.com/Emiloart/HDIP/packages/go/foundation/runtime"
	"github.com/Emiloart/HDIP/services/verifier-api/internal/config"
	"github.com/Emiloart/HDIP/services/verifier-api/internal/httpapi"
)

func Run(ctx context.Context, cfg config.Config) error {
	logger, err := observability.NewJSONLogger(cfg.ServiceName, cfg.LogLevel)
	if err != nil {
		return err
	}

	handler := httpapi.NewMux(logger, cfg)
	return runtime.RunHTTP(ctx, logger, runtime.HTTPConfig{
		Address:           cfg.Address(),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ShutdownTimeout:   cfg.ShutdownTimeout,
	}, handler)
}
