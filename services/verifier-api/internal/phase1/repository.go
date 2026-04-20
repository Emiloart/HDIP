package phase1

import (
	"context"
	"time"

	"github.com/Emiloart/HDIP/packages/go/foundation/authctx"
)

type VerificationDecision string

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

type IssuerRecord struct {
	IssuerID                  string
	DisplayName               string
	TrustState                string
	AllowedTemplateIDs        []string
	VerificationKeyReferences []string
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
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

type IssuerRecordRepository interface {
	GetIssuerRecord(ctx context.Context, issuerID string) (IssuerRecord, error)
}

type VerificationRequestRepository interface {
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
