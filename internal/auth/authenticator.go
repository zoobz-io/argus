package auth

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/zoobz-io/rocco"
)

// claims maps the Zitadel-specific JWT claims we extract.
type claims struct {
	Roles map[string]map[string]string `json:"urn:zitadel:iam:org:project:roles"`
	Email string                       `json:"email"`
	OrgID string                       `json:"urn:zitadel:iam:user:resourceowner:id"`
	Scope string                       `json:"scope"`
}

// NewAuthenticator creates a rocco-compatible authenticator function that validates
// OIDC JWTs issued by the given issuer. If audience is empty, audience verification
// is skipped (useful for development).
func NewAuthenticator(ctx context.Context, issuer, audience string) (func(context.Context, *http.Request) (rocco.Identity, error), error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery failed for %s: %w", issuer, err)
	}

	cfg := &oidc.Config{
		ClientID: audience,
	}
	if audience == "" {
		cfg.SkipClientIDCheck = true
	}

	verifier := provider.Verifier(cfg)

	return func(ctx context.Context, r *http.Request) (rocco.Identity, error) {
		token, err := extractBearerToken(r)
		if err != nil {
			return nil, err
		}

		idToken, err := verifier.Verify(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("token verification failed: %w", err)
		}

		var c claims
		if err := idToken.Claims(&c); err != nil {
			return nil, fmt.Errorf("failed to parse claims: %w", err)
		}

		return identityFromClaims(idToken.Subject, &c), nil
	}, nil
}

// identityFromClaims maps parsed Zitadel JWT claims to a ZitadelIdentity.
func identityFromClaims(subject string, c *claims) *ZitadelIdentity {
	roles := make([]string, 0, len(c.Roles))
	for role := range c.Roles {
		roles = append(roles, role)
	}
	sort.Strings(roles)

	var scopes []string
	if c.Scope != "" {
		scopes = strings.Split(c.Scope, " ")
	}

	return &ZitadelIdentity{
		sub:      subject,
		tenantID: c.OrgID,
		email:    c.Email,
		roles:    roles,
		scopes:   scopes,
	}
}

// extractBearerToken pulls the JWT from the Authorization header.
func extractBearerToken(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("missing Authorization header")
	}
	if !strings.HasPrefix(auth, "Bearer ") {
		return "", fmt.Errorf("invalid Authorization header format")
	}
	return strings.TrimPrefix(auth, "Bearer "), nil
}
