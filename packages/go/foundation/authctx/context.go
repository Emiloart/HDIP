package authctx

import (
	"errors"
	"net/http"
	"strings"
)

type ActorType string

const (
	ActorTypeIssuerOperator     ActorType = "issuer_operator"
	ActorTypeVerifierIntegrator ActorType = "verifier_integrator"
)

type Attribution struct {
	PrincipalID             string
	OrganizationID          string
	ActorType               ActorType
	Scopes                  []string
	AuthenticationReference string
}

type IssuerOperatorExtractor interface {
	IssuerOperatorFromRequest(r *http.Request) (Attribution, error)
}

type VerifierIntegratorExtractor interface {
	VerifierIntegratorFromRequest(r *http.Request) (Attribution, error)
}

func (a Attribution) Validate() error {
	switch {
	case strings.TrimSpace(a.PrincipalID) == "":
		return errors.New("principal id must not be empty")
	case strings.TrimSpace(a.OrganizationID) == "":
		return errors.New("organization id must not be empty")
	case strings.TrimSpace(string(a.ActorType)) == "":
		return errors.New("actor type must not be empty")
	case a.ActorType != ActorTypeIssuerOperator && a.ActorType != ActorTypeVerifierIntegrator:
		return errors.New("actor type must be a known phase1 actor")
	case len(a.Scopes) == 0:
		return errors.New("at least one scope is required")
	case strings.TrimSpace(a.AuthenticationReference) == "":
		return errors.New("authentication reference must not be empty")
	default:
		return nil
	}
}

func (a Attribution) ValidateFor(expected ActorType) error {
	if err := a.Validate(); err != nil {
		return err
	}

	if a.ActorType != expected {
		return errors.New("actor type does not match expected attribution boundary")
	}

	return nil
}

func (a Attribution) HasScope(scope string) bool {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		return false
	}

	for _, candidate := range a.Scopes {
		if strings.TrimSpace(candidate) == scope {
			return true
		}
	}

	return false
}
