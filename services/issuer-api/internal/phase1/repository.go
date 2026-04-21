package phase1

import (
	"context"
	"errors"
	"time"

	"github.com/Emiloart/HDIP/packages/go/foundation/authctx"
)

type CredentialStatus string

var ErrRecordNotFound = errors.New("phase1 record not found")

const (
	CredentialStatusActive     CredentialStatus = "active"
	CredentialStatusRevoked    CredentialStatus = "revoked"
	CredentialStatusSuperseded CredentialStatus = "superseded"
)

type KYCClaims struct {
	FullLegalName      string
	DateOfBirth        string
	CountryOfResidence string
	DocumentCountry    string
	KYCLevel           string
	VerifiedAt         time.Time
	ExpiresAt          time.Time
}

type CredentialArtifact struct {
	Kind      string
	MediaType string
	Value     string
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

type CredentialRecord struct {
	CredentialID             string
	IssuerID                 string
	TemplateID               string
	SubjectReference         string
	Claims                   KYCClaims
	ArtifactDigest           string
	CredentialArtifact       *CredentialArtifact
	ArtifactReference        string
	Status                   CredentialStatus
	StatusReference          string
	IssuedAt                 time.Time
	ExpiresAt                time.Time
	StatusUpdatedAt          time.Time
	SupersededByCredentialID string
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

type CredentialRecordRepository interface {
	NextCredentialID(ctx context.Context, templateID string) (string, error)
	CreateCredentialRecord(ctx context.Context, record CredentialRecord) error
	GetCredentialRecord(ctx context.Context, credentialID string) (CredentialRecord, error)
	UpdateCredentialStatus(
		ctx context.Context,
		credentialID string,
		status CredentialStatus,
		statusUpdatedAt time.Time,
		supersededByCredentialID string,
	) error
}

type AuditRecordRepository interface {
	AppendAuditRecord(ctx context.Context, record AuditRecord) error
}
