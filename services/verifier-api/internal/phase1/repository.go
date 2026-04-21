package phase1

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Emiloart/HDIP/packages/go/foundation/authctx"
)

type VerificationDecision string

var ErrRecordNotFound = errors.New("phase1 record not found")

const (
	VerificationDecisionAllow  VerificationDecision = "allow"
	VerificationDecisionDeny   VerificationDecision = "deny"
	VerificationDecisionReview VerificationDecision = "review"
)

type CredentialStatusSnapshot string

const (
	CredentialStatusSnapshotActive     CredentialStatusSnapshot = "active"
	CredentialStatusSnapshotRevoked    CredentialStatusSnapshot = "revoked"
	CredentialStatusSnapshotSuperseded CredentialStatusSnapshot = "superseded"
	CredentialStatusSnapshotExpired    CredentialStatusSnapshot = "expired"
)

type CredentialRecord struct {
	CredentialID   string
	IssuerID       string
	TemplateID     string
	ArtifactDigest string
	ExpiresAt      time.Time
	Status         CredentialStatusSnapshot
}

type IssuerRecord struct {
	IssuerID                  string
	DisplayName               string
	TrustState                string
	AllowedTemplateIDs        []string
	VerificationKeyReferences []string
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

type IssuerTrustRecord struct {
	IssuerID                  string
	TrustState                string
	AllowedTemplateIDs        []string
	VerificationKeyReferences []string
}

type VerificationRequestRecord struct {
	VerificationID            string
	VerifierID                string
	SubmittedCredentialDigest string
	CredentialID              string
	PolicyID                  string
	RequestedAt               time.Time
	Actor                     authctx.Attribution
	IdempotencyKey            string
}

type VerificationResultRecord struct {
	VerificationID   string
	IssuerID         string
	Decision         VerificationDecision
	ReasonCodes      []string
	IssuerTrustState string
	CredentialStatus CredentialStatusSnapshot
	EvaluatedAt      time.Time
	ResponseVersion  string
}

type AuditRecord struct {
	AuditID        string
	Actor          authctx.Attribution
	Action         string
	ResourceType   string
	ResourceID     string
	RequestID      string
	IdempotencyKey string
	Outcome        string
	OccurredAt     time.Time
	ServiceName    string
}

type CredentialRecordRepository interface {
	GetCredentialRecord(ctx context.Context, credentialID string) (CredentialRecord, error)
	GetCredentialRecordByArtifactDigest(ctx context.Context, artifactDigest string) (CredentialRecord, error)
}

type VerificationRequestRepository interface {
	NextVerificationID(ctx context.Context) (string, error)
	CreateVerificationRequestRecord(ctx context.Context, record VerificationRequestRecord) error
	GetVerificationRequestRecord(ctx context.Context, verificationID string) (VerificationRequestRecord, error)
}

type VerificationResultRepository interface {
	CreateVerificationResultRecord(ctx context.Context, record VerificationResultRecord) error
	GetVerificationResultRecord(ctx context.Context, verificationID string) (VerificationResultRecord, error)
}

type AuditRecordRepository interface {
	AppendAuditRecord(ctx context.Context, record AuditRecord) error
}

type IdempotencyRecord struct {
	Operation            string
	CallerPrincipalID    string
	CallerOrganizationID string
	CallerActorType      string
	IdempotencyKey       string
	RequestFingerprint   string
	ResponseStatusCode   int
	ResourceType         string
	ResourceID           string
	Location             string
	ResponseBody         json.RawMessage
	CreatedAt            time.Time
}

type IdempotencyRecordRepository interface {
	CreateIdempotencyRecord(ctx context.Context, record IdempotencyRecord) error
	GetIdempotencyRecord(
		ctx context.Context,
		operation string,
		callerOrganizationID string,
		callerPrincipalID string,
		callerActorType string,
		idempotencyKey string,
	) (IdempotencyRecord, error)
}

type TrustReadRepository interface {
	GetIssuerTrustRecord(ctx context.Context, issuerID string) (IssuerTrustRecord, error)
}
