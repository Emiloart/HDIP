package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/Emiloart/HDIP/packages/go/foundation/httpx"
	"github.com/Emiloart/HDIP/services/trust-registry/internal/config"
)

func NewMux(logger *slog.Logger, cfg config.Config) (http.Handler, error) {
	phase1Handler, err := newPhase1TrustHandler(cfg)
	if err != nil {
		return nil, err
	}

	return newMuxWithPhase1Handler(logger, cfg, phase1Handler), nil
}

func newMuxWithPhase1Handler(logger *slog.Logger, cfg config.Config, phase1Handler *phase1TrustHandler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/healthz", httpx.HealthHandler(cfg.ServiceName, cfg.BuildVersion))
	mux.Handle("/readyz", httpx.ReadyHandler(cfg.ServiceName, cfg.BuildVersion))
	mux.HandleFunc("GET /internal/v1/phase1/issuers/{issuerId}/trust", phase1Handler.getIssuerTrust)

	return httpx.Chain(
		httpx.RouteHandler(mux),
		httpx.RequestID,
		httpx.Recover(logger),
		httpx.AccessLog(logger),
		httpx.ContextTimeout(cfg.RequestTimeout),
	)
}
