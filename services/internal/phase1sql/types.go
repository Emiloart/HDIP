package phase1sql

import (
	"encoding/json"
	"time"
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
	Status                   string
	StatusReference          string
	IssuedAt                 time.Time
	ExpiresAt                time.Time
	StatusUpdatedAt          time.Time
	SupersededByCredentialID string
}

type Actor struct {
	PrincipalID             string
	OrganizationID          string
	ActorType               string
	Scopes                  []string
	AuthenticationReference string
}

type VerificationRequestRecord struct {
	VerificationID            string
	VerifierID                string
	SubmittedCredentialDigest string
	CredentialID              string
	PolicyID                  string
	RequestedAt               time.Time
	Actor                     Actor
	IdempotencyKey            string
}

type VerificationResultRecord struct {
	VerificationID   string
	IssuerID         string
	Decision         string
	ReasonCodes      []string
	IssuerTrustState string
	CredentialStatus string
	EvaluatedAt      time.Time
	ResponseVersion  string
}

type AuditRecord struct {
	AuditID        string
	Actor          Actor
	Action         string
	ResourceType   string
	ResourceID     string
	RequestID      string
	IdempotencyKey string
	Outcome        string
	OccurredAt     time.Time
	ServiceName    string
}

type IdempotencyRecord struct {
	Operation            string
	CallerPrincipalID    string
	CallerOrganizationID string
	CallerActorType      string
	IdempotencyKey       string
	RequestFingerprint   string
	State                string
	ResponseStatusCode   int
	ResourceType         string
	ResourceID           string
	Location             string
	ResponseBody         json.RawMessage
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type IdempotencyReservationResult struct {
	Outcome string
	Record  IdempotencyRecord
}

const (
	IdempotencyStateReserved  = "reserved"
	IdempotencyStateCompleted = "completed"

	IdempotencyReservationReserved   = "reserved"
	IdempotencyReservationReplay     = "replay"
	IdempotencyReservationConflict   = "conflict"
	IdempotencyReservationInProgress = "in_progress"
)
