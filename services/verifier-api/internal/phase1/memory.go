package phase1

import (
	"context"
	"strings"
	"sync"
	"time"
)

const placeholderTemplateID = "hdip-passport-basic"

type InMemoryStore struct {
	mu       sync.RWMutex
	requests map[string]VerificationRequestRecord
	results  map[string]VerificationResultRecord
	audits   []AuditRecord
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		requests: make(map[string]VerificationRequestRecord),
		results:  make(map[string]VerificationResultRecord),
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

func (s *InMemoryStore) CreateVerificationRequestRecord(_ context.Context, record VerificationRequestRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.requests[record.VerificationID] = record
	return nil
}

func (s *InMemoryStore) GetVerificationRequestRecord(_ context.Context, verificationID string) (VerificationRequestRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.requests[verificationID]
	if !ok {
		return VerificationRequestRecord{}, ErrRecordNotFound
	}

	return record, nil
}

func (s *InMemoryStore) CreateVerificationResultRecord(_ context.Context, record VerificationResultRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.results[record.VerificationID] = record
	return nil
}

func (s *InMemoryStore) GetVerificationResultRecord(_ context.Context, verificationID string) (VerificationResultRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.results[verificationID]
	if !ok {
		return VerificationResultRecord{}, ErrRecordNotFound
	}

	return record, nil
}

func (s *InMemoryStore) AppendAuditRecord(_ context.Context, record AuditRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.audits = append(s.audits, record)
	return nil
}
