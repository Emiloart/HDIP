package authctx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHeaderIssuerOperatorExtractor(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("X-HDIP-Principal-ID", "issuer_operator_alex")
	request.Header.Set("X-HDIP-Organization-ID", "issuer_org_hdip")
	request.Header.Set("X-HDIP-Auth-Reference", "session_issuer_001")
	request.Header.Set("X-HDIP-Scopes", "issuer.credentials.issue, issuer.credentials.read")

	attribution, err := (HeaderIssuerOperatorExtractor{}).IssuerOperatorFromRequest(request)
	if err != nil {
		t.Fatalf("expected extraction to succeed, got %v", err)
	}

	if attribution.ActorType != ActorTypeIssuerOperator {
		t.Fatalf("unexpected actor type: %s", attribution.ActorType)
	}

	if !attribution.HasScope("issuer.credentials.issue") {
		t.Fatal("expected extracted attribution to include issue scope")
	}
}

func TestHeaderVerifierIntegratorExtractorRejectsMissingHeaders(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", nil)

	if _, err := (HeaderVerifierIntegratorExtractor{}).VerifierIntegratorFromRequest(request); err == nil {
		t.Fatal("expected missing attribution headers to fail")
	}
}

func TestRequireScope(t *testing.T) {
	attribution := Attribution{
		PrincipalID:             "verifier_integrator_alpha",
		OrganizationID:          "verifier_org_alpha",
		ActorType:               ActorTypeVerifierIntegrator,
		Scopes:                  []string{"verifier.requests.create"},
		AuthenticationReference: "credential_verifier_001",
	}

	if err := RequireScope(attribution, "verifier.requests.create"); err != nil {
		t.Fatalf("expected scope check to pass, got %v", err)
	}

	if err := RequireScope(attribution, "verifier.results.read"); err == nil {
		t.Fatal("expected missing scope to fail")
	}
}
