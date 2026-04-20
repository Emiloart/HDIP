package httpx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Emiloart/HDIP/packages/go/foundation/testutil"
)

func TestHealthHandlerMatchesCanonicalExample(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	HealthHandler("example-service", "test").ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	testutil.AssertJSONMatchesFixture(t, recorder.Body.Bytes(), "schemas/examples/common/health.ok.json")
}

func TestReadyHandlerMatchesCanonicalExample(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	ReadyHandler("example-service", "test").ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	testutil.AssertJSONMatchesFixture(t, recorder.Body.Bytes(), "schemas/examples/common/health.ready.json")
}

func TestWriteErrorMatchesCanonicalExample(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), requestIDContextKey{}, "req-1")

	WriteError(recorder, ctx, http.StatusNotFound, "route_not_found", "route not found")

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}

	testutil.AssertJSONMatchesFixture(t, recorder.Body.Bytes(), "schemas/examples/common/error-envelope.route-not-found.json")
}
