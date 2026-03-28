// Package auth provides OIDC-based authentication for the application.
package auth

import "github.com/zoobz-io/rocco"

// Compile-time assertion.
var _ rocco.Identity = (*ZitadelIdentity)(nil)

// ZitadelIdentity maps Zitadel OIDC JWT claims to rocco's Identity interface.
type ZitadelIdentity struct {
	sub      string
	tenantID string
	email    string
	roles    []string
	scopes   []string
}

// ID returns the Zitadel user subject.
func (z *ZitadelIdentity) ID() string { return z.sub }

// TenantID returns the Zitadel organization ID.
func (z *ZitadelIdentity) TenantID() string { return z.tenantID }

// Email returns the user's email address.
func (z *ZitadelIdentity) Email() string { return z.email }

// Scopes returns the granted scopes.
func (z *ZitadelIdentity) Scopes() []string { return z.scopes }

// Roles returns the granted roles.
func (z *ZitadelIdentity) Roles() []string { return z.roles }

// HasScope checks if the identity has a specific scope.
func (z *ZitadelIdentity) HasScope(s string) bool {
	for _, sc := range z.scopes {
		if sc == s {
			return true
		}
	}
	return false
}

// HasRole checks if the identity has a specific role.
func (z *ZitadelIdentity) HasRole(r string) bool {
	for _, rl := range z.roles {
		if rl == r {
			return true
		}
	}
	return false
}

// Stats returns nil — rate limiting stats are not tracked on the identity.
func (z *ZitadelIdentity) Stats() map[string]int { return nil }
