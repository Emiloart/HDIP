package phase1

import (
	"context"
	"strings"
	"sync"
	"time"
)

const placeholderTemplateID = "hdip-passport-basic"

type InMemoryStore struct {
	mu          sync.RWMutex
	credentials map[string]CredentialRecord
	audits      []AuditRecord
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		credentials: make(map[string]CredentialRecord),
	}
}

func (s *InMemoryStore) GetIssuerRecord(_ context.Context, issuerID string) (IssuerRecord, error) {
	issuerID = strings.TrimSpace(issuerID)
	if issuerID == "" {
		return IssuerRecord{}, ErrRecordNotFound
	}

	now := time.Date(2026, time.April, 20, 9, 0, 0, 0, time.UTC)
	return IssuerRecord{
		IssuerID:                  issuerID,
		DisplayName:               "HDIP Passport Issuer",
		TrustState:                "active",
		AllowedTemplateIDs:        []string{placeholderTemplateID},
		VerificationKeyReferences: []string{"key:issuer.hdip.dev:2026-04"},
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}, nil
}

func (s *InMemoryStore) CreateCredentialRecord(_ context.Context, record CredentialRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.credentials[record.CredentialID] = record
	return nil
}

func (s *InMemoryStore) GetCredentialRecord(_ context.Context, credentialID string) (CredentialRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.credentials[credentialID]
	if !ok {
		return CredentialRecord{}, ErrRecordNotFound
	}

	return record, nil
}

func (s *InMemoryStore) UpdateCredentialStatus(
	_ context.Context,
	credentialID string,
	status CredentialStatus,
	statusUpdatedAt time.Time,
	supersededByCredentialID string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.credentials[credentialID]
	if !ok {
		return ErrRecordNotFound
	}

	record.Status = status
	record.StatusUpdatedAt = statusUpdatedAt
	record.SupersededByCredentialID = supersededByCredentialID
	s.credentials[credentialID] = record

	return nil
}

func (s *InMemoryStore) AppendAuditRecord(_ context.Context, record AuditRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.audits = append(s.audits, record)
	return nil
}
