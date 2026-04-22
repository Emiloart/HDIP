package phase1sql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const TrustBootstrapAction = "trust-registry.phase1.bootstrap.apply"

type TrustBootstrapDocument struct {
	Issuers []TrustBootstrapIssuerRecord `json:"issuers"`
}

type TrustBootstrapIssuerRecord struct {
	IssuerID                  string   `json:"issuerId"`
	DisplayName               string   `json:"displayName"`
	TrustState                string   `json:"trustState"`
	AllowedTemplateIDs        []string `json:"allowedTemplateIds"`
	VerificationKeyReferences []string `json:"verificationKeyReferences"`
}

type TrustBootstrapResult struct {
	Applied int
}

func ApplyTrustBootstrapFile(ctx context.Context, store *Store, path string, now time.Time) (TrustBootstrapResult, error) {
	if store == nil {
		return TrustBootstrapResult{}, errors.New("phase1 sql store is required")
	}

	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return TrustBootstrapResult{}, nil
	}

	raw, err := os.ReadFile(trimmedPath)
	if err != nil {
		return TrustBootstrapResult{}, fmt.Errorf("read phase1 trust bootstrap file: %w", err)
	}

	var document TrustBootstrapDocument
	if err := json.Unmarshal(raw, &document); err != nil {
		return TrustBootstrapResult{}, fmt.Errorf("decode phase1 trust bootstrap file: %w", err)
	}

	return ApplyTrustBootstrapDocument(ctx, store, filepath.Base(trimmedPath), document, now)
}

func ApplyTrustBootstrapDocument(
	ctx context.Context,
	store *Store,
	source string,
	document TrustBootstrapDocument,
	now time.Time,
) (TrustBootstrapResult, error) {
	if store == nil {
		return TrustBootstrapResult{}, errors.New("phase1 sql store is required")
	}

	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	sourceLabel := strings.TrimSpace(source)
	if sourceLabel == "" {
		sourceLabel = "inline"
	}

	result := TrustBootstrapResult{}
	for index, bootstrapIssuer := range document.Issuers {
		record, err := bootstrapIssuer.toIssuerRecord()
		if err != nil {
			return TrustBootstrapResult{}, fmt.Errorf("bootstrap issuer %d: %w", index, err)
		}

		existingRecord, err := store.GetIssuerRecord(ctx, record.IssuerID)
		switch {
		case err == nil:
			record.CreatedAt = existingRecord.CreatedAt
		case errors.Is(err, ErrRecordNotFound):
			record.CreatedAt = now
		default:
			return TrustBootstrapResult{}, fmt.Errorf("load existing issuer trust record: %w", err)
		}
		record.UpdatedAt = now

		if err := store.UpsertIssuerRecord(ctx, record); err != nil {
			return TrustBootstrapResult{}, fmt.Errorf("upsert issuer trust record: %w", err)
		}

		if err := store.AppendAuditRecord(ctx, AuditRecord{
			AuditID:      fmt.Sprintf("trust-bootstrap:%s:%s:%d:%d", sourceLabel, record.IssuerID, index, now.UnixNano()),
			Actor:        trustBootstrapActor(sourceLabel),
			Action:       TrustBootstrapAction,
			ResourceType: "issuer_trust_record",
			ResourceID:   record.IssuerID,
			Outcome:      "succeeded",
			OccurredAt:   now,
			ServiceName:  "trust-registry",
		}); err != nil {
			return TrustBootstrapResult{}, fmt.Errorf("append trust bootstrap audit record: %w", err)
		}

		result.Applied++
	}

	return result, nil
}

func (r TrustBootstrapIssuerRecord) toIssuerRecord() (IssuerRecord, error) {
	switch {
	case strings.TrimSpace(r.IssuerID) == "":
		return IssuerRecord{}, errors.New("issuerId must not be empty")
	case strings.TrimSpace(r.DisplayName) == "":
		return IssuerRecord{}, errors.New("displayName must not be empty")
	case strings.TrimSpace(r.TrustState) == "":
		return IssuerRecord{}, errors.New("trustState must not be empty")
	default:
		return IssuerRecord{
			IssuerID:                  strings.TrimSpace(r.IssuerID),
			DisplayName:               strings.TrimSpace(r.DisplayName),
			TrustState:                strings.TrimSpace(r.TrustState),
			AllowedTemplateIDs:        append([]string(nil), r.AllowedTemplateIDs...),
			VerificationKeyReferences: append([]string(nil), r.VerificationKeyReferences...),
		}, nil
	}
}

func trustBootstrapActor(source string) Actor {
	return Actor{
		PrincipalID:             "trust-registry-bootstrap",
		OrganizationID:          "trust-registry",
		ActorType:               "service",
		Scopes:                  []string{"trust.runtime.bootstrap"},
		AuthenticationReference: "bootstrap:" + source,
	}
}
