package httpapi

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Emiloart/HDIP/packages/go/foundation/httpx"
	"github.com/Emiloart/HDIP/services/trust-registry/internal/config"
	phase1 "github.com/Emiloart/HDIP/services/trust-registry/internal/phase1"
)

type phase1TrustHandler struct {
	store             *phase1.RuntimeStore
	internalAuthToken string
}

type issuerTrustPayload struct {
	IssuerID                  string   `json:"issuerId"`
	TrustState                string   `json:"trustState"`
	AllowedTemplateIDs        []string `json:"allowedTemplateIds"`
	VerificationKeyReferences []string `json:"verificationKeyReferences"`
}

func newPhase1TrustHandler(cfg config.Config) (*phase1TrustHandler, error) {
	store, err := phase1.OpenStore(phase1.StoreOptions{
		DatabaseDriver:  cfg.Phase1DatabaseDriver,
		DatabaseURL:     cfg.Phase1DatabaseURL,
		LegacyStatePath: cfg.Phase1StatePath,
	})
	if err != nil {
		return nil, err
	}

	if _, err := phase1.ApplyBootstrapFile(context.Background(), store, cfg.TrustBootstrapPath, time.Now().UTC()); err != nil {
		_ = store.Close()
		return nil, err
	}

	return newPhase1TrustHandlerWithStoreAndToken(store, cfg.InternalAuthToken), nil
}

func newPhase1TrustHandlerWithStore(store *phase1.RuntimeStore) *phase1TrustHandler {
	return newPhase1TrustHandlerWithStoreAndToken(store, "")
}

func newPhase1TrustHandlerWithStoreAndToken(store *phase1.RuntimeStore, internalAuthToken string) *phase1TrustHandler {
	return &phase1TrustHandler{
		store:             store,
		internalAuthToken: strings.TrimSpace(internalAuthToken),
	}
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
	token := strings.TrimSpace(strings.TrimPrefix(authorizationHeader, "Bearer "))
	if !strings.HasPrefix(authorizationHeader, "Bearer ") ||
		token == "" ||
		subtle.ConstantTimeCompare([]byte(token), []byte(h.internalAuthToken)) != 1 {
		httpx.WriteError(w, r.Context(), http.StatusUnauthorized, "unauthenticated", "internal trust runtime credentials are required")
		return false
	}

	return true
}
