package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/Emiloart/HDIP/packages/go/foundation/httpx"
	"github.com/Emiloart/HDIP/services/verifier-api/internal/config"
)

func NewMux(logger *slog.Logger, cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/healthz", httpx.HealthHandler(cfg.ServiceName, cfg.BuildVersion))
	mux.Handle("/readyz", httpx.ReadyHandler(cfg.ServiceName, cfg.BuildVersion))

	return httpx.Chain(
		httpx.RouteHandler(mux),
		httpx.RequestID,
		httpx.Recover(logger),
		httpx.AccessLog(logger),
		httpx.ContextTimeout(cfg.RequestTimeout),
	)
}
