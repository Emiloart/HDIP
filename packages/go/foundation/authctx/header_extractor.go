package authctx

import (
	"errors"
	"net/http"
	"strings"
)

const (
	headerPrincipalID             = "X-HDIP-Principal-ID"
	headerOrganizationID          = "X-HDIP-Organization-ID"
	headerAuthenticationReference = "X-HDIP-Auth-Reference"
	headerScopes                  = "X-HDIP-Scopes"
)

type HeaderIssuerOperatorExtractor struct{}
type HeaderVerifierIntegratorExtractor struct{}

func (HeaderIssuerOperatorExtractor) IssuerOperatorFromRequest(r *http.Request) (Attribution, error) {
	attribution := headerAttributionFromRequest(r, ActorTypeIssuerOperator)
	if err := attribution.ValidateFor(ActorTypeIssuerOperator); err != nil {
		return Attribution{}, err
	}

	return attribution, nil
}

func (HeaderVerifierIntegratorExtractor) VerifierIntegratorFromRequest(r *http.Request) (Attribution, error) {
	attribution := headerAttributionFromRequest(r, ActorTypeVerifierIntegrator)
	if err := attribution.ValidateFor(ActorTypeVerifierIntegrator); err != nil {
		return Attribution{}, err
	}

	return attribution, nil
}

func headerAttributionFromRequest(r *http.Request, actorType ActorType) Attribution {
	return Attribution{
		PrincipalID:             strings.TrimSpace(r.Header.Get(headerPrincipalID)),
		OrganizationID:          strings.TrimSpace(r.Header.Get(headerOrganizationID)),
		ActorType:               actorType,
		Scopes:                  parseScopes(r.Header.Get(headerScopes)),
		AuthenticationReference: strings.TrimSpace(r.Header.Get(headerAuthenticationReference)),
	}
}

func parseScopes(raw string) []string {
	parts := strings.Split(raw, ",")
	scopes := make([]string, 0, len(parts))
	for _, part := range parts {
		scope := strings.TrimSpace(part)
		if scope == "" {
			continue
		}

		scopes = append(scopes, scope)
	}

	return scopes
}

var ErrMissingScope = errors.New("required scope missing")

func RequireScope(attribution Attribution, scope string) error {
	if !attribution.HasScope(scope) {
		return ErrMissingScope
	}

	return nil
}
