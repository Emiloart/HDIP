package httpx

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONBodyRejectsUnknownFields(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	request := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(strings.NewReader(`{"name":"alex","extra":"field"}`)))

	var decoded payload
	if err := DecodeJSONBody(request, &decoded); err == nil {
		t.Fatal("expected unknown field decode to fail")
	}
}

func TestDecodeJSONBodyRejectsMultipleValues(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	request := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(strings.NewReader(`{"name":"alex"} {"name":"second"}`)))

	var decoded payload
	if err := DecodeJSONBody(request, &decoded); err == nil {
		t.Fatal("expected multiple json values to fail")
	}
}

func TestDecodeJSONBodyAcceptsSingleValue(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	request := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(strings.NewReader(`{"name":"alex"}`)))

	var decoded payload
	if err := DecodeJSONBody(request, &decoded); err != nil {
		t.Fatalf("expected decode to succeed, got %v", err)
	}

	if decoded.Name != "alex" {
		t.Fatalf("unexpected decoded payload: %+v", decoded)
	}
}
