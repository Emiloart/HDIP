package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/Emiloart/HDIP/packages/go/foundation/httpx"
	"github.com/Emiloart/HDIP/services/verifier-api/internal/config"
)

func NewMux(logger *slog.Logger, cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	phase1Handler := newPhase1VerifierHandler()
	mux.Handle("/healthz", httpx.HealthHandler(cfg.ServiceName, cfg.BuildVersion))
	mux.Handle("/readyz", httpx.ReadyHandler(cfg.ServiceName, cfg.BuildVersion))
	mux.HandleFunc("GET /v1/verifier/policy-requests/{policyId}", func(w http.ResponseWriter, r *http.Request) {
		policyID := r.PathValue("policyId")
		policy, ok := stubPolicyRequest(policyID)
		if !ok {
			httpx.WriteError(w, r.Context(), http.StatusNotFound, "policy_not_found", "verifier policy request not found")
			return
		}

		httpx.WriteJSON(w, http.StatusOK, policy)
	})
	mux.HandleFunc("POST /v1/verifier/verifications", phase1Handler.createVerification)
	mux.HandleFunc("GET /v1/verifier/verifications/{verificationId}", phase1Handler.getVerification)
	mux.HandleFunc("GET /v1/verifier/results/{requestId}/stub", func(w http.ResponseWriter, r *http.Request) {
		requestID := r.PathValue("requestId")
		result, ok := stubVerifierResult(requestID)
		if !ok {
			httpx.WriteError(w, r.Context(), http.StatusNotFound, "result_not_found", "stub verifier result not found")
			return
		}

		httpx.WriteJSON(w, http.StatusOK, result)
	})

	return httpx.Chain(
		httpx.RouteHandler(mux),
		httpx.RequestID,
		httpx.Recover(logger),
		httpx.AccessLog(logger),
		httpx.ContextTimeout(cfg.RequestTimeout),
	)
}
