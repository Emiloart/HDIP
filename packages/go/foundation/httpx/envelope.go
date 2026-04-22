package httpx

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type ErrorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId,omitempty"`
}

type ErrorEnvelope struct {
	Error ErrorDetail `json:"error"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

func WriteJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if payload == nil {
		return
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(true)
	_ = encoder.Encode(payload)
}

func WriteError(w http.ResponseWriter, ctx context.Context, statusCode int, code string, message string) {
	WriteJSON(w, statusCode, ErrorEnvelope{
		Error: ErrorDetail{
			Code:      code,
			Message:   message,
			RequestID: RequestIDFromContext(ctx),
		},
	})
}

func HealthHandler(service string, version string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, HealthResponse{
			Status:  "ok",
			Service: service,
			Version: version,
		})
	})
}

func ReadyHandler(service string, version string) http.Handler {
	return ReadyHandlerWithCheck(service, version, "", nil)
}

func ReadyHandlerWithCheck(
	service string,
	version string,
	runtimeMode string,
	check func(context.Context) error,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if normalizedRuntimeMode := strings.TrimSpace(runtimeMode); normalizedRuntimeMode != "" {
			w.Header().Set("X-HDIP-Phase1-Runtime-Mode", normalizedRuntimeMode)
		}
		if check != nil {
			if err := check(r.Context()); err != nil {
				WriteError(w, r.Context(), http.StatusServiceUnavailable, "not_ready", err.Error())
				return
			}
		}

		WriteJSON(w, http.StatusOK, HealthResponse{
			Status:  "ready",
			Service: service,
			Version: version,
		})
	})
}
