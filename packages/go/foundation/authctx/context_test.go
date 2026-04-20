package authctx

import "testing"

func TestAttributionValidateForIssuerOperator(t *testing.T) {
	attribution := Attribution{
		PrincipalID:             "issuer_operator_alex",
		OrganizationID:          "issuer_org_hdip",
		ActorType:               ActorTypeIssuerOperator,
		Scopes:                  []string{"issuer.credentials.issue", "issuer.credentials.read"},
		AuthenticationReference: "session_issuer_001",
	}

	if err := attribution.ValidateFor(ActorTypeIssuerOperator); err != nil {
		t.Fatalf("expected issuer operator attribution to validate, got %v", err)
	}
}

func TestAttributionValidateRejectsMismatchedActorType(t *testing.T) {
	attribution := Attribution{
		PrincipalID:             "issuer_operator_alex",
		OrganizationID:          "issuer_org_hdip",
		ActorType:               ActorTypeVerifierIntegrator,
		Scopes:                  []string{"issuer.credentials.issue"},
		AuthenticationReference: "session_issuer_001",
	}

	if err := attribution.ValidateFor(ActorTypeIssuerOperator); err == nil {
		t.Fatal("expected mismatched actor type to fail validation")
	}
}

func TestAttributionValidateRejectsMissingScopes(t *testing.T) {
	attribution := Attribution{
		PrincipalID:             "verifier_integrator_alpha",
		OrganizationID:          "verifier_org_alpha",
		ActorType:               ActorTypeVerifierIntegrator,
		Scopes:                  nil,
		AuthenticationReference: "credential_verifier_001",
	}

	if err := attribution.Validate(); err == nil {
		t.Fatal("expected empty scope list to fail validation")
	}
}

func TestAttributionHasScope(t *testing.T) {
	attribution := Attribution{
		PrincipalID:             "verifier_integrator_alpha",
		OrganizationID:          "verifier_org_alpha",
		ActorType:               ActorTypeVerifierIntegrator,
		Scopes:                  []string{"verifier.requests.create", "verifier.results.read"},
		AuthenticationReference: "credential_verifier_001",
	}

	if !attribution.HasScope("verifier.results.read") {
		t.Fatal("expected attribution to report granted scope")
	}

	if attribution.HasScope("issuer.credentials.issue") {
		t.Fatal("expected attribution to reject unrelated scope")
	}
}
