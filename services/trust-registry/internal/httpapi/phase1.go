package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Emiloart/HDIP/packages/go/foundation/httpx"
	"github.com/Emiloart/HDIP/services/trust-registry/internal/config"
	phase1 "github.com/Emiloart/HDIP/services/trust-registry/internal/phase1"
)

type phase1TrustHandler struct {
	store      *phase1.RuntimeStore
	authorizer internalAuthorizer
}

type issuerTrustPayload struct {
	IssuerID                  string   `json:"issuerId"`
	TrustState                string   `json:"trustState"`
	AllowedTemplateIDs        []string `json:"allowedTemplateIds"`
	VerificationKeyReferences []string `json:"verificationKeyReferences"`
}

func newPhase1TrustHandler(cfg config.Config) (*phase1TrustHandler, error) {
	store, err := phase1.OpenStore(phase1.StoreOptions{
		RuntimeMode:           cfg.Phase1RuntimeMode,
		DatabaseDriver:        cfg.Phase1DatabaseDriver,
		DatabaseURL:           cfg.Phase1DatabaseURL,
		TransitionalStatePath: cfg.Phase1StatePath,
	})
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(cfg.Phase1RuntimeMode) == phase1.RuntimeModeTransitionalJSON {
		if _, err := phase1.ApplyBootstrapFile(context.Background(), store, cfg.TrustBootstrapPath, time.Now().UTC()); err != nil {
			_ = store.Close()
			return nil, err
		}
	} else if strings.TrimSpace(cfg.TrustBootstrapPath) != "" {
		_ = store.Close()
		return nil, fmt.Errorf("trust registry bootstrap path must be applied through the phase1sql CLI when the primary sql path is enabled")
	}

	authorizer, err := newHydraIntrospectionAuthorizer(
		cfg.TrustRuntimeHydraIntrospectionURL,
		cfg.TrustRuntimeHydraClientID,
		cfg.TrustRuntimeHydraClientSecret,
		cfg.TrustRuntimeHydraExpectedClientID,
		cfg.TrustRuntimeHydraRequiredScope,
		&http.Client{Timeout: cfg.RequestTimeout},
	)
	if err != nil {
		_ = store.Close()
		return nil, err
	}

	return newPhase1TrustHandlerWithStoreAndAuthorizer(store, authorizer), nil
}

func newPhase1TrustHandlerWithStore(store *phase1.RuntimeStore) *phase1TrustHandler {
	return newPhase1TrustHandlerWithStoreAndAuthorizer(store, staticInternalAuthorizer{principal: internalPrincipal{ClientID: "verifier-api"}})
}

func newPhase1TrustHandlerWithStoreAndAuthorizer(store *phase1.RuntimeStore, authorizer internalAuthorizer) *phase1TrustHandler {
	if authorizer == nil {
		authorizer = staticInternalAuthorizer{principal: internalPrincipal{ClientID: "verifier-api"}}
	}

	return &phase1TrustHandler{
		store:      store,
		authorizer: authorizer,
	}
}

func (h *phase1TrustHandler) readiness(ctx context.Context) error {
	if err := h.store.CheckReadiness(ctx, true); err != nil {
		return err
	}
	if h.authorizer != nil {
		if err := h.authorizer.Check(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (h *phase1TrustHandler) runtimeMode() string {
	return h.store.RuntimeMode()
}

func (h *phase1TrustHandler) getIssuerTrust(w http.ResponseWriter, r *http.Request) {
	if !h.requireInternalAuth(w, r) {
		return
	}

	record, err := h.store.GetIssuerRecord(r.Context(), r.PathValue("issuerId"))
	if err != nil {
		if errors.Is(err, phase1.ErrRecordNotFound) {
			httpx.WriteError(w, r.Context(), http.StatusNotFound, "issuer_not_found", "issuer trust record not found")
			return
		}

		httpx.WriteError(w, r.Context(), http.StatusInternalServerError, "persistence_error", "issuer trust record could not be loaded")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, issuerTrustPayload{
		IssuerID:                  record.IssuerID,
		TrustState:                record.TrustState,
		AllowedTemplateIDs:        append([]string(nil), record.AllowedTemplateIDs...),
		VerificationKeyReferences: append([]string(nil), record.VerificationKeyReferences...),
	})
}

func (h *phase1TrustHandler) requireInternalAuth(w http.ResponseWriter, r *http.Request) bool {
	authorizationHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(authorizationHeader, "Bearer ") {
		httpx.WriteError(w, r.Context(), http.StatusUnauthorized, "unauthenticated", "internal trust runtime credentials are required")
		return false
	}

	token := strings.TrimSpace(strings.TrimPrefix(authorizationHeader, "Bearer "))
	if token == "" {
		httpx.WriteError(w, r.Context(), http.StatusUnauthorized, "unauthenticated", "internal trust runtime credentials are required")
		return false
	}

	_, err := h.authorizer.Authorize(r.Context(), token)
	switch {
	case err == nil:
		return true
	case errors.Is(err, ErrInternalAuthUnauthenticated):
		httpx.WriteError(w, r.Context(), http.StatusUnauthorized, "unauthenticated", "internal trust runtime credentials are required")
	case errors.Is(err, ErrInternalAuthForbidden):
		httpx.WriteError(w, r.Context(), http.StatusForbidden, "forbidden", "internal trust runtime credentials do not authorize this client")
	default:
		httpx.WriteError(w, r.Context(), http.StatusServiceUnavailable, "trust_runtime_auth_unavailable", "internal trust runtime authorization is unavailable")
	}

	return false
}

type staticInternalAuthorizer struct {
	err       error
	principal internalPrincipal
}

func (a staticInternalAuthorizer) Authorize(context.Context, string) (internalPrincipal, error) {
	if a.err != nil {
		return internalPrincipal{}, a.err
	}

	return a.principal, nil
}

func (a staticInternalAuthorizer) Check(context.Context) error {
	return a.err
}
