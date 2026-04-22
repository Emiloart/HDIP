package phase1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	phase1sql "github.com/Emiloart/HDIP/services/internal/phase1sql"
)

const bootstrapAction = "trust-registry.phase1.bootstrap.apply"

type BootstrapDocument struct {
	Issuers []BootstrapIssuerRecord `json:"issuers"`
}

type BootstrapIssuerRecord struct {
	IssuerID                  string   `json:"issuerId"`
	DisplayName               string   `json:"displayName"`
	TrustState                string   `json:"trustState"`
	AllowedTemplateIDs        []string `json:"allowedTemplateIds"`
	VerificationKeyReferences []string `json:"verificationKeyReferences"`
}

type BootstrapResult struct {
	Applied int
}

func ApplyBootstrapFile(ctx context.Context, store *RuntimeStore, path string, now time.Time) (BootstrapResult, error) {
	if store == nil {
		return BootstrapResult{}, errors.New("runtime store is required")
	}
	if store.sql != nil {
		result, err := phase1sql.ApplyTrustBootstrapFile(ctx, store.sql, path, now)
		if err != nil {
			return BootstrapResult{}, err
		}

		return BootstrapResult{Applied: result.Applied}, nil
	}

	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return BootstrapResult{}, nil
	}

	raw, err := os.ReadFile(trimmedPath)
	if err != nil {
		return BootstrapResult{}, fmt.Errorf("read phase1 trust bootstrap file: %w", err)
	}

	var document BootstrapDocument
	if err := json.Unmarshal(raw, &document); err != nil {
		return BootstrapResult{}, fmt.Errorf("decode phase1 trust bootstrap file: %w", err)
	}

	return ApplyBootstrapDocument(ctx, store, filepath.Base(trimmedPath), document, now)
}

func ApplyBootstrapDocument(ctx context.Context, store *RuntimeStore, source string, document BootstrapDocument, now time.Time) (BootstrapResult, error) {
	if store == nil {
		return BootstrapResult{}, errors.New("runtime store is required")
	}
	if store.sql != nil {
		result, err := phase1sql.ApplyTrustBootstrapDocument(ctx, store.sql, source, toSQLBootstrapDocument(document), now)
		if err != nil {
			return BootstrapResult{}, err
		}

		return BootstrapResult{Applied: result.Applied}, nil
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

	result := BootstrapResult{}
	for index, bootstrapIssuer := range document.Issuers {
		record, err := bootstrapIssuer.toIssuerRecord()
		if err != nil {
			return BootstrapResult{}, fmt.Errorf("bootstrap issuer %d: %w", index, err)
		}

		existing, err := store.GetIssuerRecord(ctx, record.IssuerID)
		switch {
		case err == nil:
			record.CreatedAt = existing.CreatedAt
		case errors.Is(err, ErrRecordNotFound):
			record.CreatedAt = now
		default:
			return BootstrapResult{}, fmt.Errorf("load existing issuer trust record: %w", err)
		}
		record.UpdatedAt = now

		if err := store.UpsertIssuerRecord(ctx, record); err != nil {
			return BootstrapResult{}, fmt.Errorf("upsert issuer trust record: %w", err)
		}

		if err := store.AppendAuditRecord(ctx, AuditRecord{
			AuditID:      fmt.Sprintf("trust-bootstrap:%s:%s:%d:%d", sourceLabel, record.IssuerID, index, now.UnixNano()),
			Actor:        bootstrapActor(sourceLabel),
			Action:       bootstrapAction,
			ResourceType: "issuer_trust_record",
			ResourceID:   record.IssuerID,
			Outcome:      "succeeded",
			OccurredAt:   now,
			ServiceName:  "trust-registry",
		}); err != nil {
			return BootstrapResult{}, fmt.Errorf("append trust bootstrap audit record: %w", err)
		}

		result.Applied++
	}

	return result, nil
}

func (r BootstrapIssuerRecord) toIssuerRecord() (IssuerRecord, error) {
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

func bootstrapActor(source string) Actor {
	return Actor{
		PrincipalID:             "trust-registry-bootstrap",
		OrganizationID:          "trust-registry",
		ActorType:               "service",
		Scopes:                  []string{"trust.runtime.bootstrap"},
		AuthenticationReference: "bootstrap:" + source,
	}
}

func toSQLBootstrapDocument(document BootstrapDocument) phase1sql.TrustBootstrapDocument {
	issuers := make([]phase1sql.TrustBootstrapIssuerRecord, 0, len(document.Issuers))
	for _, issuer := range document.Issuers {
		issuers = append(issuers, phase1sql.TrustBootstrapIssuerRecord{
			IssuerID:                  issuer.IssuerID,
			DisplayName:               issuer.DisplayName,
			TrustState:                issuer.TrustState,
			AllowedTemplateIDs:        append([]string(nil), issuer.AllowedTemplateIDs...),
			VerificationKeyReferences: append([]string(nil), issuer.VerificationKeyReferences...),
		})
	}

	return phase1sql.TrustBootstrapDocument{Issuers: issuers}
}
