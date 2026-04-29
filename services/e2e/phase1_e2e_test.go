package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	phase1sql "github.com/Emiloart/HDIP/services/internal/phase1sql"
)

const (
	issuerID   = "did:web:issuer.hdip.dev"
	templateID = "hdip-passport-basic"
	policyID   = "kyc-passport-basic"
)

type credentialArtifact struct {
	Kind      string `json:"kind"`
	MediaType string `json:"mediaType"`
	Value     string `json:"value"`
}

type issuanceResponse struct {
	CredentialID       string             `json:"credentialId"`
	CredentialArtifact credentialArtifact `json:"credentialArtifact"`
}

type verificationResult struct {
	VerificationID   string   `json:"verificationId"`
	Decision         string   `json:"decision"`
	ReasonCodes      []string `json:"reasonCodes"`
	CredentialStatus string   `json:"credentialStatus"`
}

type sandboxConfig struct {
	DatabaseDriver                    string
	DatabaseURL                       string
	HydraAdminURL                     string
	HydraPublicURL                    string
	VerifierTrustClientID             string
	VerifierTrustClientSecret         string
	TrustRegistryIntrospectionID      string
	TrustRegistryIntrospectionSecret  string
	TrustRegistryIntrospectionScope   string
	TrustRegistryIntrospectionBaseURL string
}

func TestPhase1SandboxLifecycle(t *testing.T) {
	if os.Getenv("HDIP_PHASE1_E2E") != "1" {
		t.Skip("set HDIP_PHASE1_E2E=1 with real SQL and Hydra to run the Phase 1 sandbox E2E test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cfg := loadSandboxConfig(t)
	root := repoRoot(t)

	if err := phase1sql.MigrateUp(ctx, cfg.DatabaseDriver, cfg.DatabaseURL); err != nil {
		t.Fatalf("migrate phase1 sql: %v", err)
	}

	store, err := phase1sql.Open(cfg.DatabaseDriver, cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("open phase1 sql store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	applyTrustBootstrap(t, ctx, store, "active")
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		applyTrustBootstrap(t, cleanupCtx, store, "active")
	})

	trustURL := startService(t, ctx, root, "trust-registry", freePort(t), map[string]string{
		"HDIP_PHASE1_DATABASE_DRIVER":                          cfg.DatabaseDriver,
		"HDIP_PHASE1_DATABASE_URL":                             cfg.DatabaseURL,
		"HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_URL":           cfg.TrustRegistryIntrospectionBaseURL + "/admin/oauth2/introspect",
		"HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_CLIENT_ID":     cfg.TrustRegistryIntrospectionID,
		"HDIP_TRUST_RUNTIME_HYDRA_INTROSPECTION_CLIENT_SECRET": cfg.TrustRegistryIntrospectionSecret,
		"HDIP_TRUST_RUNTIME_HYDRA_EXPECTED_CLIENT_ID":          cfg.VerifierTrustClientID,
		"HDIP_TRUST_RUNTIME_HYDRA_REQUIRED_SCOPE":              cfg.TrustRegistryIntrospectionScope,
	})
	issuerURL := startService(t, ctx, root, "issuer-api", freePort(t), map[string]string{
		"HDIP_PHASE1_DATABASE_DRIVER": cfg.DatabaseDriver,
		"HDIP_PHASE1_DATABASE_URL":    cfg.DatabaseURL,
	})
	verifierURL := startService(t, ctx, root, "verifier-api", freePort(t), map[string]string{
		"HDIP_PHASE1_DATABASE_DRIVER":            cfg.DatabaseDriver,
		"HDIP_PHASE1_DATABASE_URL":               cfg.DatabaseURL,
		"HDIP_TRUST_REGISTRY_BASE_URL":           trustURL,
		"HDIP_TRUST_RUNTIME_HYDRA_TOKEN_URL":     cfg.HydraPublicURL + "/oauth2/token",
		"HDIP_TRUST_RUNTIME_HYDRA_CLIENT_ID":     cfg.VerifierTrustClientID,
		"HDIP_TRUST_RUNTIME_HYDRA_CLIENT_SECRET": cfg.VerifierTrustClientSecret,
		"HDIP_TRUST_RUNTIME_HYDRA_SCOPE":         cfg.TrustRegistryIntrospectionScope,
	})

	waitReady(t, ctx, trustURL)
	waitReady(t, ctx, issuerURL)
	waitReady(t, ctx, verifierURL)

	issued := issueCredential(t, ctx, issuerURL)
	allowed := verifyCredential(t, ctx, verifierURL, issued, "allow")

	revokeCredential(t, ctx, issuerURL, issued.CredentialID)
	revokedDenied := verifyCredential(t, ctx, verifierURL, issued, "deny")
	if !slices.Contains(revokedDenied.ReasonCodes, "credential_status_revoked") {
		t.Fatalf("expected revoked verification reason, got %v", revokedDenied.ReasonCodes)
	}

	applyTrustBootstrap(t, ctx, store, "suspended")
	suspendedDenied := verifyCredential(t, ctx, verifierURL, issued, "deny")
	if !slices.Contains(suspendedDenied.ReasonCodes, "issuer_suspended") {
		t.Fatalf("expected issuer_suspended reason, got %v", suspendedDenied.ReasonCodes)
	}

	records, err := store.ListAuditRecords(ctx)
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}
	assertAuditRecord(t, records, "issuer.credentials.issue", issued.CredentialID)
	assertAuditRecord(t, records, "issuer.credentials.status.write", issued.CredentialID)
	assertAuditRecord(t, records, "verifier.requests.create", allowed.VerificationID)
	assertAuditRecord(t, records, "verifier.requests.create", revokedDenied.VerificationID)
}

func loadSandboxConfig(t *testing.T) sandboxConfig {
	t.Helper()

	databaseURL := firstNonEmpty(os.Getenv("HDIP_PHASE1_E2E_DATABASE_URL"), os.Getenv("DATABASE_URL"))
	cfg := sandboxConfig{
		DatabaseDriver:                   firstNonEmpty(os.Getenv("DATABASE_DRIVER"), "pgx"),
		DatabaseURL:                      databaseURL,
		HydraAdminURL:                    trimTrailingSlash(os.Getenv("HYDRA_ADMIN_URL")),
		HydraPublicURL:                   trimTrailingSlash(os.Getenv("HYDRA_PUBLIC_URL")),
		VerifierTrustClientID:            firstNonEmpty(os.Getenv("VERIFIER_TRUST_CLIENT_ID"), "verifier-api"),
		VerifierTrustClientSecret:        os.Getenv("VERIFIER_TRUST_CLIENT_SECRET"),
		TrustRegistryIntrospectionID:     firstNonEmpty(os.Getenv("TRUST_REGISTRY_INTROSPECTION_CLIENT_ID"), "trust-registry"),
		TrustRegistryIntrospectionSecret: os.Getenv("TRUST_REGISTRY_INTROSPECTION_CLIENT_SECRET"),
		TrustRegistryIntrospectionScope:  firstNonEmpty(os.Getenv("TRUST_RUNTIME_SCOPE"), "trust.runtime.read"),
	}
	cfg.TrustRegistryIntrospectionBaseURL = cfg.HydraAdminURL

	required := map[string]string{
		"DATABASE_URL or HDIP_PHASE1_E2E_DATABASE_URL": databaseURL,
		"HYDRA_ADMIN_URL":                            cfg.HydraAdminURL,
		"HYDRA_PUBLIC_URL":                           cfg.HydraPublicURL,
		"VERIFIER_TRUST_CLIENT_SECRET":               cfg.VerifierTrustClientSecret,
		"TRUST_REGISTRY_INTROSPECTION_CLIENT_SECRET": cfg.TrustRegistryIntrospectionSecret,
	}
	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			t.Fatalf("%s is required for HDIP_PHASE1_E2E=1", name)
		}
	}

	return cfg
}

func repoRoot(t *testing.T) string {
	t.Helper()

	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	root := filepath.Clean(filepath.Join(workingDirectory, "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err != nil {
		t.Fatalf("resolve repo root from %s: %v", workingDirectory, err)
	}

	return root
}

func applyTrustBootstrap(t *testing.T, ctx context.Context, store *phase1sql.Store, trustState string) {
	t.Helper()

	_, err := phase1sql.ApplyTrustBootstrapDocument(ctx, store, "phase1-e2e", phase1sql.TrustBootstrapDocument{
		Issuers: []phase1sql.TrustBootstrapIssuerRecord{
			{
				IssuerID:                  issuerID,
				DisplayName:               "HDIP Sandbox Issuer",
				TrustState:                trustState,
				AllowedTemplateIDs:        []string{templateID},
				VerificationKeyReferences: []string{"phase1-opaque-artifact-reference"},
			},
		},
	}, time.Now().UTC())
	if err != nil {
		t.Fatalf("apply %s trust bootstrap: %v", trustState, err)
	}
}

func startService(t *testing.T, ctx context.Context, root string, serviceName string, port int, env map[string]string) string {
	t.Helper()

	serviceDir := filepath.Join(root, "services", serviceName)
	logFile := filepath.Join(t.TempDir(), serviceName+".log")
	logHandle, err := os.Create(logFile)
	if err != nil {
		t.Fatalf("create %s log file: %v", serviceName, err)
	}
	t.Cleanup(func() {
		_ = logHandle.Close()
	})

	command := exec.CommandContext(ctx, "go", "run", "./cmd/"+serviceName)
	command.Dir = serviceDir
	command.Stdout = logHandle
	command.Stderr = logHandle
	command.Env = append(os.Environ(),
		"HDIP_HOST=127.0.0.1",
		fmt.Sprintf("HDIP_PORT=%d", port),
	)
	for key, value := range env {
		command.Env = append(command.Env, key+"="+value)
	}

	if err := command.Start(); err != nil {
		t.Fatalf("start %s: %v", serviceName, err)
	}

	t.Cleanup(func() {
		if command.Process != nil {
			_ = command.Process.Kill()
		}
		_ = command.Wait()

		if t.Failed() {
			raw, _ := os.ReadFile(logFile)
			t.Logf("%s log:\n%s", serviceName, string(raw))
		}
	})

	return fmt.Sprintf("http://127.0.0.1:%d", port)
}

func waitReady(t *testing.T, ctx context.Context, baseURL string) {
	t.Helper()

	deadline := time.Now().Add(45 * time.Second)
	client := &http.Client{Timeout: time.Second}
	for time.Now().Before(deadline) {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/readyz", nil)
		if err != nil {
			t.Fatalf("build ready request: %v", err)
		}

		response, err := client.Do(request)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	t.Fatalf("%s did not become ready", baseURL)
}

func issueCredential(t *testing.T, ctx context.Context, issuerURL string) issuanceResponse {
	t.Helper()

	payload := map[string]any{
		"templateId":       templateID,
		"subjectReference": "phase1-e2e-subject",
		"claims": map[string]string{
			"fullLegalName":      "Phase One Sandbox",
			"dateOfBirth":        "1990-01-02",
			"countryOfResidence": "NG",
			"documentCountry":    "NG",
			"kycLevel":           "basic",
			"verifiedAt":         "2026-04-28T10:00:00Z",
			"expiresAt":          "2099-01-01T00:00:00Z",
		},
	}

	var response issuanceResponse
	doJSON(t, ctx, http.MethodPost, issuerURL+"/v1/issuer/credentials", payload, http.StatusCreated, issuerHeaders(), &response)
	if response.CredentialID == "" || response.CredentialArtifact.Value == "" {
		t.Fatalf("issuance response missing credential ID or artifact: %+v", response)
	}

	return response
}

func verifyCredential(t *testing.T, ctx context.Context, verifierURL string, issued issuanceResponse, expectedDecision string) verificationResult {
	t.Helper()

	payload := map[string]any{
		"policyId":           policyID,
		"credentialId":       issued.CredentialID,
		"credentialArtifact": issued.CredentialArtifact,
	}

	var response verificationResult
	doJSON(t, ctx, http.MethodPost, verifierURL+"/v1/verifier/verifications", payload, http.StatusCreated, verifierHeaders(), &response)
	if response.Decision != expectedDecision {
		t.Fatalf("verification returned %q, expected %q: %+v", response.Decision, expectedDecision, response)
	}

	return response
}

func revokeCredential(t *testing.T, ctx context.Context, issuerURL string, credentialID string) {
	t.Helper()

	payload := map[string]string{"status": "revoked"}
	var response map[string]any
	doJSON(
		t,
		ctx,
		http.MethodPost,
		issuerURL+"/v1/issuer/credentials/"+credentialID+"/status",
		payload,
		http.StatusOK,
		issuerHeaders(),
		&response,
	)
}

func doJSON(
	t *testing.T,
	ctx context.Context,
	method string,
	url string,
	payload any,
	expectedStatus int,
	headers map[string]string,
	target any,
) {
	t.Helper()

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	request, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(rawPayload))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", fmt.Sprintf("phase1-e2e-%d", time.Now().UnixNano()))
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer response.Body.Close()

	rawResponse, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if response.StatusCode != expectedStatus {
		t.Fatalf("%s %s returned %d, expected %d: %s", method, url, response.StatusCode, expectedStatus, string(rawResponse))
	}
	if err := json.Unmarshal(rawResponse, target); err != nil {
		t.Fatalf("decode response: %v: %s", err, string(rawResponse))
	}
}

func issuerHeaders() map[string]string {
	return map[string]string{
		"X-HDIP-Principal-ID":    "issuer_operator_e2e",
		"X-HDIP-Organization-ID": issuerID,
		"X-HDIP-Auth-Reference":  "phase1-e2e",
		"X-HDIP-Scopes":          "issuer.credentials.issue, issuer.credentials.read, issuer.credentials.status.write",
	}
}

func verifierHeaders() map[string]string {
	return map[string]string{
		"X-HDIP-Principal-ID":    "verifier_integrator_e2e",
		"X-HDIP-Organization-ID": "verifier_org_e2e",
		"X-HDIP-Auth-Reference":  "phase1-e2e",
		"X-HDIP-Scopes":          "verifier.requests.create, verifier.results.read",
	}
}

func assertAuditRecord(t *testing.T, records []phase1sql.AuditRecord, action string, resourceID string) {
	t.Helper()

	for _, record := range records {
		if record.Action == action && record.ResourceID == resourceID {
			return
		}
	}

	t.Fatalf("missing audit record action=%s resourceID=%s", action, resourceID)
}

func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("allocate free port: %v", err)
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}

func trimTrailingSlash(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}
