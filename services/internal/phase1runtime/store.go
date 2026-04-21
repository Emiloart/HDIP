package phase1runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var ErrRecordNotFound = errors.New("phase1 runtime record not found")

var inMemoryCounter int64
var storeLocks sync.Map

type Store struct {
	path string
	lock *sync.Mutex
}

type persistedState struct {
	Sequences            map[string]int                       `json:"sequences"`
	Issuers              map[string]IssuerRecord              `json:"issuers"`
	Credentials          map[string]CredentialRecord          `json:"credentials"`
	VerificationRequests map[string]VerificationRequestRecord `json:"verificationRequests"`
	VerificationResults  map[string]VerificationResultRecord  `json:"verificationResults"`
	IdempotencyRecords   map[string]IdempotencyRecord         `json:"idempotencyRecords"`
	Audits               []AuditRecord                        `json:"audits"`
}

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
	Operation            string          `json:"operation"`
	CallerPrincipalID    string          `json:"callerPrincipalId"`
	CallerOrganizationID string          `json:"callerOrganizationId"`
	CallerActorType      string          `json:"callerActorType"`
	IdempotencyKey       string          `json:"idempotencyKey"`
	RequestFingerprint   string          `json:"requestFingerprint"`
	State                string          `json:"state"`
	ResponseStatusCode   int             `json:"responseStatusCode"`
	ResourceType         string          `json:"resourceType"`
	ResourceID           string          `json:"resourceId"`
	Location             string          `json:"location,omitempty"`
	ResponseBody         json.RawMessage `json:"responseBody"`
	CreatedAt            time.Time       `json:"createdAt"`
	UpdatedAt            time.Time       `json:"updatedAt"`
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

func Open(path string) (*Store, error) {
	canonicalPath, err := canonicalStorePath(path)
	if err != nil {
		return nil, err
	}

	lock := pathLock(canonicalPath)
	store := &Store{
		path: canonicalPath,
		lock: lock,
	}

	lock.Lock()
	defer lock.Unlock()

	if err := ensureStoreFile(canonicalPath); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) Close() error {
	return nil
}

func (s *Store) NextCredentialID(ctx context.Context, templateID string) (string, error) {
	_ = ctx

	var credentialID string
	err := s.withMutableState(func(state *persistedState) error {
		sequence := nextSequence(state, "credential")
		credentialID = formatCredentialID(templateID, sequence)
		return nil
	})
	if err != nil {
		return "", err
	}

	return credentialID, nil
}

func (s *Store) NextVerificationID(ctx context.Context) (string, error) {
	_ = ctx

	var verificationID string
	err := s.withMutableState(func(state *persistedState) error {
		sequence := nextSequence(state, "verification")
		verificationID = fmt.Sprintf("verification_hdip_%03d", sequence)
		return nil
	})
	if err != nil {
		return "", err
	}

	return verificationID, nil
}

func (s *Store) UpsertIssuerRecord(ctx context.Context, record IssuerRecord) error {
	_ = ctx

	return s.withMutableState(func(state *persistedState) error {
		state.Issuers[strings.TrimSpace(record.IssuerID)] = cloneIssuerRecord(record)
		return nil
	})
}

func (s *Store) DeleteIssuerRecord(ctx context.Context, issuerID string) error {
	_ = ctx

	return s.withMutableState(func(state *persistedState) error {
		normalizedIssuerID := strings.TrimSpace(issuerID)
		if _, ok := state.Issuers[normalizedIssuerID]; !ok {
			return ErrRecordNotFound
		}

		delete(state.Issuers, normalizedIssuerID)
		return nil
	})
}

func (s *Store) GetIssuerRecord(ctx context.Context, issuerID string) (IssuerRecord, error) {
	_ = ctx

	var record IssuerRecord
	err := s.withReadState(func(state persistedState) error {
		storedRecord, ok := state.Issuers[strings.TrimSpace(issuerID)]
		if !ok {
			return ErrRecordNotFound
		}

		record = cloneIssuerRecord(storedRecord)
		return nil
	})
	if err != nil {
		return IssuerRecord{}, err
	}

	return record, nil
}

func (s *Store) CreateCredentialRecord(ctx context.Context, record CredentialRecord) error {
	_ = ctx

	return s.withMutableState(func(state *persistedState) error {
		if _, ok := state.Credentials[record.CredentialID]; ok {
			return fmt.Errorf("credential record already exists: %s", record.CredentialID)
		}

		state.Credentials[record.CredentialID] = cloneCredentialRecord(record)
		return nil
	})
}

func (s *Store) UpsertCredentialRecord(ctx context.Context, record CredentialRecord) error {
	_ = ctx

	return s.withMutableState(func(state *persistedState) error {
		state.Credentials[record.CredentialID] = cloneCredentialRecord(record)
		return nil
	})
}

func (s *Store) GetCredentialRecord(ctx context.Context, credentialID string) (CredentialRecord, error) {
	_ = ctx

	var record CredentialRecord
	err := s.withReadState(func(state persistedState) error {
		storedRecord, ok := state.Credentials[strings.TrimSpace(credentialID)]
		if !ok {
			return ErrRecordNotFound
		}

		record = cloneCredentialRecord(storedRecord)
		return nil
	})
	if err != nil {
		return CredentialRecord{}, err
	}

	return record, nil
}

func (s *Store) GetCredentialRecordByArtifactDigest(ctx context.Context, artifactDigest string) (CredentialRecord, error) {
	_ = ctx

	var record CredentialRecord
	err := s.withReadState(func(state persistedState) error {
		normalizedArtifactDigest := strings.TrimSpace(artifactDigest)
		for _, candidate := range state.Credentials {
			if strings.TrimSpace(candidate.ArtifactDigest) == normalizedArtifactDigest {
				record = cloneCredentialRecord(candidate)
				return nil
			}
		}

		return ErrRecordNotFound
	})
	if err != nil {
		return CredentialRecord{}, err
	}

	return record, nil
}

func (s *Store) UpdateCredentialStatus(
	ctx context.Context,
	credentialID string,
	status string,
	statusUpdatedAt time.Time,
	supersededByCredentialID string,
) error {
	_ = ctx

	return s.withMutableState(func(state *persistedState) error {
		normalizedCredentialID := strings.TrimSpace(credentialID)
		record, ok := state.Credentials[normalizedCredentialID]
		if !ok {
			return ErrRecordNotFound
		}

		record.Status = status
		record.StatusUpdatedAt = statusUpdatedAt.UTC()
		record.SupersededByCredentialID = strings.TrimSpace(supersededByCredentialID)
		state.Credentials[normalizedCredentialID] = cloneCredentialRecord(record)
		return nil
	})
}

func (s *Store) CreateVerificationRequestRecord(ctx context.Context, record VerificationRequestRecord) error {
	_ = ctx

	return s.withMutableState(func(state *persistedState) error {
		if _, ok := state.VerificationRequests[record.VerificationID]; ok {
			return fmt.Errorf("verification request already exists: %s", record.VerificationID)
		}

		state.VerificationRequests[record.VerificationID] = cloneVerificationRequestRecord(record)
		return nil
	})
}

func (s *Store) GetVerificationRequestRecord(ctx context.Context, verificationID string) (VerificationRequestRecord, error) {
	_ = ctx

	var record VerificationRequestRecord
	err := s.withReadState(func(state persistedState) error {
		storedRecord, ok := state.VerificationRequests[strings.TrimSpace(verificationID)]
		if !ok {
			return ErrRecordNotFound
		}

		record = cloneVerificationRequestRecord(storedRecord)
		return nil
	})
	if err != nil {
		return VerificationRequestRecord{}, err
	}

	return record, nil
}

func (s *Store) CreateVerificationResultRecord(ctx context.Context, record VerificationResultRecord) error {
	_ = ctx

	return s.withMutableState(func(state *persistedState) error {
		if _, ok := state.VerificationResults[record.VerificationID]; ok {
			return fmt.Errorf("verification result already exists: %s", record.VerificationID)
		}

		state.VerificationResults[record.VerificationID] = cloneVerificationResultRecord(record)
		return nil
	})
}

func (s *Store) GetVerificationResultRecord(ctx context.Context, verificationID string) (VerificationResultRecord, error) {
	_ = ctx

	var record VerificationResultRecord
	err := s.withReadState(func(state persistedState) error {
		storedRecord, ok := state.VerificationResults[strings.TrimSpace(verificationID)]
		if !ok {
			return ErrRecordNotFound
		}

		record = cloneVerificationResultRecord(storedRecord)
		return nil
	})
	if err != nil {
		return VerificationResultRecord{}, err
	}

	return record, nil
}

func (s *Store) AppendAuditRecord(ctx context.Context, record AuditRecord) error {
	_ = ctx

	return s.withMutableState(func(state *persistedState) error {
		state.Audits = append(state.Audits, cloneAuditRecord(record))
		return nil
	})
}

func (s *Store) ListAuditRecords(ctx context.Context) ([]AuditRecord, error) {
	_ = ctx

	var records []AuditRecord
	err := s.withReadState(func(state persistedState) error {
		records = make([]AuditRecord, 0, len(state.Audits))
		for _, record := range state.Audits {
			records = append(records, cloneAuditRecord(record))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (s *Store) CreateIdempotencyRecord(ctx context.Context, record IdempotencyRecord) error {
	_ = ctx

	return s.withMutableState(func(state *persistedState) error {
		key := idempotencyStorageKey(record.Operation, record.CallerOrganizationID, record.CallerPrincipalID, record.CallerActorType, record.IdempotencyKey)
		if _, ok := state.IdempotencyRecords[key]; ok {
			return fmt.Errorf("idempotency record already exists: %s", key)
		}

		state.IdempotencyRecords[key] = cloneIdempotencyRecord(record)
		return nil
	})
}

func (s *Store) ReserveIdempotencyRecord(ctx context.Context, record IdempotencyRecord) (IdempotencyReservationResult, error) {
	_ = ctx

	var result IdempotencyReservationResult
	err := s.withMutableState(func(state *persistedState) error {
		key := idempotencyStorageKey(record.Operation, record.CallerOrganizationID, record.CallerPrincipalID, record.CallerActorType, record.IdempotencyKey)
		storedRecord, ok := state.IdempotencyRecords[key]
		if !ok {
			reservedRecord := cloneIdempotencyRecord(record)
			now := reservedRecord.CreatedAt.UTC()
			if now.IsZero() {
				now = time.Now().UTC()
			}
			reservedRecord.CreatedAt = now
			reservedRecord.UpdatedAt = now
			reservedRecord.State = IdempotencyStateReserved
			state.IdempotencyRecords[key] = reservedRecord
			result = IdempotencyReservationResult{
				Outcome: IdempotencyReservationReserved,
				Record:  cloneIdempotencyRecord(reservedRecord),
			}
			return nil
		}

		if strings.TrimSpace(storedRecord.RequestFingerprint) != strings.TrimSpace(record.RequestFingerprint) {
			result = IdempotencyReservationResult{
				Outcome: IdempotencyReservationConflict,
				Record:  cloneIdempotencyRecord(storedRecord),
			}
			return nil
		}

		if strings.TrimSpace(storedRecord.State) == IdempotencyStateCompleted {
			result = IdempotencyReservationResult{
				Outcome: IdempotencyReservationReplay,
				Record:  cloneIdempotencyRecord(storedRecord),
			}
			return nil
		}

		result = IdempotencyReservationResult{
			Outcome: IdempotencyReservationInProgress,
			Record:  cloneIdempotencyRecord(storedRecord),
		}
		return nil
	})
	if err != nil {
		return IdempotencyReservationResult{}, err
	}

	return result, nil
}

func (s *Store) GetIdempotencyRecord(
	ctx context.Context,
	operation string,
	callerOrganizationID string,
	callerPrincipalID string,
	callerActorType string,
	idempotencyKey string,
) (IdempotencyRecord, error) {
	_ = ctx

	var record IdempotencyRecord
	err := s.withReadState(func(state persistedState) error {
		key := idempotencyStorageKey(operation, callerOrganizationID, callerPrincipalID, callerActorType, idempotencyKey)
		storedRecord, ok := state.IdempotencyRecords[key]
		if !ok {
			return ErrRecordNotFound
		}

		record = cloneIdempotencyRecord(storedRecord)
		return nil
	})
	if err != nil {
		return IdempotencyRecord{}, err
	}

	return record, nil
}

func (s *Store) ListIdempotencyRecords(ctx context.Context) ([]IdempotencyRecord, error) {
	_ = ctx

	var records []IdempotencyRecord
	err := s.withReadState(func(state persistedState) error {
		records = make([]IdempotencyRecord, 0, len(state.IdempotencyRecords))
		for _, record := range state.IdempotencyRecords {
			records = append(records, cloneIdempotencyRecord(record))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (s *Store) CompleteIdempotencyRecord(ctx context.Context, record IdempotencyRecord) error {
	_ = ctx

	return s.withMutableState(func(state *persistedState) error {
		key := idempotencyStorageKey(record.Operation, record.CallerOrganizationID, record.CallerPrincipalID, record.CallerActorType, record.IdempotencyKey)
		storedRecord, ok := state.IdempotencyRecords[key]
		if !ok {
			return ErrRecordNotFound
		}

		storedRecord.State = IdempotencyStateCompleted
		storedRecord.ResponseStatusCode = record.ResponseStatusCode
		storedRecord.ResourceType = record.ResourceType
		storedRecord.ResourceID = record.ResourceID
		storedRecord.Location = record.Location
		storedRecord.ResponseBody = append(json.RawMessage(nil), record.ResponseBody...)
		storedRecord.UpdatedAt = record.UpdatedAt.UTC()
		if storedRecord.UpdatedAt.IsZero() {
			storedRecord.UpdatedAt = time.Now().UTC()
		}
		state.IdempotencyRecords[key] = cloneIdempotencyRecord(storedRecord)
		return nil
	})
}

func (s *Store) ReleaseIdempotencyRecord(
	ctx context.Context,
	operation string,
	callerOrganizationID string,
	callerPrincipalID string,
	callerActorType string,
	idempotencyKey string,
) error {
	_ = ctx

	return s.withMutableState(func(state *persistedState) error {
		key := idempotencyStorageKey(operation, callerOrganizationID, callerPrincipalID, callerActorType, idempotencyKey)
		storedRecord, ok := state.IdempotencyRecords[key]
		if !ok {
			return nil
		}

		if strings.TrimSpace(storedRecord.State) == IdempotencyStateCompleted {
			return nil
		}

		delete(state.IdempotencyRecords, key)
		return nil
	})
}

func (s *Store) withReadState(fn func(state persistedState) error) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	state, err := loadState(s.path)
	if err != nil {
		return err
	}

	return fn(state)
}

func (s *Store) withMutableState(fn func(state *persistedState) error) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	state, err := loadState(s.path)
	if err != nil {
		return err
	}

	if err := fn(&state); err != nil {
		return err
	}

	return saveState(s.path, state)
}

func loadState(path string) (persistedState, error) {
	raw, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return defaultState(), nil
	case err != nil:
		return persistedState{}, fmt.Errorf("read phase1 runtime store: %w", err)
	case len(raw) == 0:
		return defaultState(), nil
	}

	var state persistedState
	if err := json.Unmarshal(raw, &state); err != nil {
		return persistedState{}, fmt.Errorf("decode phase1 runtime store: %w", err)
	}

	normalizeState(&state)
	return state, nil
}

func saveState(path string, state persistedState) error {
	normalizeState(&state)

	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode phase1 runtime store: %w", err)
	}

	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, raw, 0o600); err != nil {
		return fmt.Errorf("write phase1 runtime store: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace phase1 runtime store: %w", err)
	}

	return nil
}

func ensureStoreFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat phase1 runtime store: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create phase1 runtime directory: %w", err)
	}

	return saveState(path, defaultState())
}

func defaultState() persistedState {
	return persistedState{
		Sequences:            make(map[string]int),
		Issuers:              make(map[string]IssuerRecord),
		Credentials:          make(map[string]CredentialRecord),
		VerificationRequests: make(map[string]VerificationRequestRecord),
		VerificationResults:  make(map[string]VerificationResultRecord),
		IdempotencyRecords:   make(map[string]IdempotencyRecord),
		Audits:               []AuditRecord{},
	}
}

func normalizeState(state *persistedState) {
	if state.Sequences == nil {
		state.Sequences = make(map[string]int)
	}
	if state.Issuers == nil {
		state.Issuers = make(map[string]IssuerRecord)
	}
	if state.Credentials == nil {
		state.Credentials = make(map[string]CredentialRecord)
	}
	if state.VerificationRequests == nil {
		state.VerificationRequests = make(map[string]VerificationRequestRecord)
	}
	if state.VerificationResults == nil {
		state.VerificationResults = make(map[string]VerificationResultRecord)
	}
	if state.IdempotencyRecords == nil {
		state.IdempotencyRecords = make(map[string]IdempotencyRecord)
	}
	if state.Audits == nil {
		state.Audits = []AuditRecord{}
	}
}

func nextSequence(state *persistedState, name string) int {
	sequence := state.Sequences[name] + 1
	state.Sequences[name] = sequence
	return sequence
}

func canonicalStorePath(path string) (string, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return "", errors.New("phase1 runtime path must not be empty")
	}

	if trimmedPath == ":memory:" {
		return filepath.Join(
			os.TempDir(),
			fmt.Sprintf("hdip-phase1-state-memory-%d.json", atomic.AddInt64(&inMemoryCounter, 1)),
		), nil
	}

	absolutePath, err := filepath.Abs(trimmedPath)
	if err != nil {
		return "", fmt.Errorf("resolve phase1 runtime path: %w", err)
	}

	return filepath.Clean(absolutePath), nil
}

func pathLock(path string) *sync.Mutex {
	lock, _ := storeLocks.LoadOrStore(path, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func cloneIssuerRecord(record IssuerRecord) IssuerRecord {
	return IssuerRecord{
		IssuerID:                  record.IssuerID,
		DisplayName:               record.DisplayName,
		TrustState:                record.TrustState,
		AllowedTemplateIDs:        append([]string(nil), record.AllowedTemplateIDs...),
		VerificationKeyReferences: append([]string(nil), record.VerificationKeyReferences...),
		CreatedAt:                 record.CreatedAt.UTC(),
		UpdatedAt:                 record.UpdatedAt.UTC(),
	}
}

func cloneCredentialArtifact(record *CredentialArtifact) *CredentialArtifact {
	if record == nil {
		return nil
	}

	return &CredentialArtifact{
		Kind:      record.Kind,
		MediaType: record.MediaType,
		Value:     record.Value,
	}
}

func cloneCredentialRecord(record CredentialRecord) CredentialRecord {
	return CredentialRecord{
		CredentialID:             record.CredentialID,
		IssuerID:                 record.IssuerID,
		TemplateID:               record.TemplateID,
		SubjectReference:         record.SubjectReference,
		Claims:                   record.Claims,
		ArtifactDigest:           record.ArtifactDigest,
		CredentialArtifact:       cloneCredentialArtifact(record.CredentialArtifact),
		ArtifactReference:        record.ArtifactReference,
		Status:                   record.Status,
		StatusReference:          record.StatusReference,
		IssuedAt:                 record.IssuedAt.UTC(),
		ExpiresAt:                record.ExpiresAt.UTC(),
		StatusUpdatedAt:          record.StatusUpdatedAt.UTC(),
		SupersededByCredentialID: record.SupersededByCredentialID,
	}
}

func cloneActor(record Actor) Actor {
	return Actor{
		PrincipalID:             record.PrincipalID,
		OrganizationID:          record.OrganizationID,
		ActorType:               record.ActorType,
		Scopes:                  append([]string(nil), record.Scopes...),
		AuthenticationReference: record.AuthenticationReference,
	}
}

func cloneVerificationRequestRecord(record VerificationRequestRecord) VerificationRequestRecord {
	return VerificationRequestRecord{
		VerificationID:            record.VerificationID,
		VerifierID:                record.VerifierID,
		SubmittedCredentialDigest: record.SubmittedCredentialDigest,
		CredentialID:              record.CredentialID,
		PolicyID:                  record.PolicyID,
		RequestedAt:               record.RequestedAt.UTC(),
		Actor:                     cloneActor(record.Actor),
		IdempotencyKey:            record.IdempotencyKey,
	}
}

func cloneVerificationResultRecord(record VerificationResultRecord) VerificationResultRecord {
	return VerificationResultRecord{
		VerificationID:   record.VerificationID,
		IssuerID:         record.IssuerID,
		Decision:         record.Decision,
		ReasonCodes:      append([]string(nil), record.ReasonCodes...),
		IssuerTrustState: record.IssuerTrustState,
		CredentialStatus: record.CredentialStatus,
		EvaluatedAt:      record.EvaluatedAt.UTC(),
		ResponseVersion:  record.ResponseVersion,
	}
}

func cloneAuditRecord(record AuditRecord) AuditRecord {
	return AuditRecord{
		AuditID:        record.AuditID,
		Actor:          cloneActor(record.Actor),
		Action:         record.Action,
		ResourceType:   record.ResourceType,
		ResourceID:     record.ResourceID,
		RequestID:      record.RequestID,
		IdempotencyKey: record.IdempotencyKey,
		Outcome:        record.Outcome,
		OccurredAt:     record.OccurredAt.UTC(),
		ServiceName:    record.ServiceName,
	}
}

func cloneIdempotencyRecord(record IdempotencyRecord) IdempotencyRecord {
	return IdempotencyRecord{
		Operation:            record.Operation,
		CallerPrincipalID:    record.CallerPrincipalID,
		CallerOrganizationID: record.CallerOrganizationID,
		CallerActorType:      record.CallerActorType,
		IdempotencyKey:       record.IdempotencyKey,
		RequestFingerprint:   record.RequestFingerprint,
		State:                record.State,
		ResponseStatusCode:   record.ResponseStatusCode,
		ResourceType:         record.ResourceType,
		ResourceID:           record.ResourceID,
		Location:             record.Location,
		ResponseBody:         append(json.RawMessage(nil), record.ResponseBody...),
		CreatedAt:            record.CreatedAt.UTC(),
		UpdatedAt:            record.UpdatedAt.UTC(),
	}
}

func idempotencyStorageKey(
	operation string,
	callerOrganizationID string,
	callerPrincipalID string,
	callerActorType string,
	idempotencyKey string,
) string {
	return strings.Join([]string{
		strings.TrimSpace(operation),
		strings.TrimSpace(callerOrganizationID),
		strings.TrimSpace(callerPrincipalID),
		strings.TrimSpace(callerActorType),
		strings.TrimSpace(idempotencyKey),
	}, "|")
}

func formatCredentialID(templateID string, sequence int) string {
	sanitizedTemplateID := strings.NewReplacer("-", "_", ".", "_").Replace(strings.TrimSpace(templateID))
	if sanitizedTemplateID == "" {
		sanitizedTemplateID = "hdip_passport_basic"
	}

	return fmt.Sprintf("cred_%s_%03d", sanitizedTemplateID, sequence)
}
