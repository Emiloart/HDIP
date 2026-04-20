package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/Emiloart/HDIP/packages/go/foundation/httpx"
	"github.com/Emiloart/HDIP/services/issuer-api/internal/config"
)

func NewMux(logger *slog.Logger, cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/healthz", httpx.HealthHandler(cfg.ServiceName, cfg.BuildVersion))
	mux.Handle("/readyz", httpx.ReadyHandler(cfg.ServiceName, cfg.BuildVersion))
	mux.HandleFunc("GET /v1/issuer/profile", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, stubIssuerProfile(cfg))
	})
	mux.HandleFunc("GET /v1/issuer/templates/{templateId}", func(w http.ResponseWriter, r *http.Request) {
		templateID := r.PathValue("templateId")
		template, ok := stubCredentialTemplate(templateID)
		if !ok {
			httpx.WriteError(w, r.Context(), http.StatusNotFound, "template_not_found", "credential template not found")
			return
		}

		httpx.WriteJSON(w, http.StatusOK, template)
	})

	return httpx.Chain(
		httpx.RouteHandler(mux),
		httpx.RequestID,
		httpx.Recover(logger),
		httpx.AccessLog(logger),
		httpx.ContextTimeout(cfg.RequestTimeout),
	)
}
