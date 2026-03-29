package handlers

import (
	"encoding/json"
	"strings"

	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/internal/audit"
	"github.com/zoobz-io/argus/internal/oauth"
	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/argus/provider"
)

// providerAuthURL returns the OAuth authorization URL for a provider.
var providerAuthURL = rocco.GET[rocco.NoBody, wire.AuthURLResponse]("/providers/{id}/auth-url", func(r *rocco.Request[rocco.NoBody]) (wire.AuthURLResponse, error) {
	tid := tenantID(r.Identity)
	providerID := pathID(r.Params, "id")
	redirectURI := r.Params.Query["redirect_uri"]

	if redirectURI == "" {
		return wire.AuthURLResponse{}, ErrMissingRedirectURI
	}

	// Validate redirect URI against allowed list.
	allowedURIs := sum.MustUse[*AllowedRedirectURIs](r)
	if !allowedURIs.IsAllowed(redirectURI) {
		return wire.AuthURLResponse{}, ErrInvalidRedirectURI
	}

	// Tenant-scoped provider lookup.
	store := sum.MustUse[contracts.Providers](r)
	p, err := store.GetProviderByTenant(r, providerID, tid)
	if err != nil {
		return wire.AuthURLResponse{}, ErrProviderNotFound
	}

	// Resolve provider implementation.
	registry := sum.MustUse[*provider.Registry](r)
	impl, err := registry.Get(string(p.Type))
	if err != nil {
		return wire.AuthURLResponse{}, ErrProviderTypeNotSupported
	}

	// Generate CSRF state.
	signer := sum.MustUse[*oauth.StateSigner](r)
	state, err := signer.Sign(providerID, tid)
	if err != nil {
		return wire.AuthURLResponse{}, err
	}

	url, err := impl.AuthURL(r, redirectURI, state)
	if err != nil {
		return wire.AuthURLResponse{}, err
	}

	return wire.AuthURLResponse{URL: url}, nil
}).
	WithSummary("Get OAuth authorization URL").
	WithTags("providers").
	WithPathParams("id").
	WithQueryParams("redirect_uri").
	WithAuthentication().
	WithErrors(ErrProviderNotFound, ErrMissingRedirectURI, ErrInvalidRedirectURI, ErrProviderTypeNotSupported)

// providerConnect exchanges an OAuth code or stores static credentials.
var providerConnect = rocco.POST[wire.ConnectRequest, wire.ConnectResponse]("/providers/{id}/connect", func(r *rocco.Request[wire.ConnectRequest]) (wire.ConnectResponse, error) {
	tid := tenantID(r.Identity)
	providerID := pathID(r.Params, "id")

	// Tenant-scoped provider lookup.
	store := sum.MustUse[contracts.Providers](r)
	p, err := store.GetProviderByTenant(r, providerID, tid)
	if err != nil {
		return wire.ConnectResponse{}, ErrProviderNotFound
	}

	var credentialsJSON string

	if isOAuthProvider(p.Type) {
		// Verify CSRF state.
		signer := sum.MustUse[*oauth.StateSigner](r)
		if stateErr := signer.Verify(r.Body.State, providerID, tid); stateErr != nil {
			return wire.ConnectResponse{}, ErrInvalidState
		}

		// Exchange code for tokens.
		registry := sum.MustUse[*provider.Registry](r)
		impl, implErr := registry.Get(string(p.Type))
		if implErr != nil {
			return wire.ConnectResponse{}, ErrProviderTypeNotSupported
		}

		creds, exchangeErr := impl.Exchange(r, r.Body.Code, r.Body.RedirectURI)
		if exchangeErr != nil {
			return wire.ConnectResponse{}, ErrOAuthExchangeFailed
		}

		credsBytes, marshalErr := json.Marshal(creds)
		if marshalErr != nil {
			return wire.ConnectResponse{}, marshalErr
		}
		credentialsJSON = string(credsBytes)
	} else {
		// Static credentials — code field carries the credentials JSON directly.
		credentialsJSON = r.Body.Code
	}

	if updateErr := store.UpdateProviderCredentials(r, providerID, credentialsJSON); updateErr != nil {
		return wire.ConnectResponse{}, updateErr
	}

	audit.Emit(r, "provider.connected", "provider", providerID, tid, r.Identity.ID(), map[string]any{
		"provider_type": string(p.Type),
	})
	return wire.ConnectResponse{Status: "connected"}, nil
}).
	WithSummary("Connect provider").
	WithTags("providers").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrProviderNotFound, ErrInvalidState, ErrProviderTypeNotSupported, ErrOAuthExchangeFailed)

// isOAuthProvider returns true for provider types that use OAuth2.
func isOAuthProvider(t models.ProviderType) bool {
	switch t {
	case models.ProviderGoogleDrive, models.ProviderOneDrive, models.ProviderDropbox:
		return true
	default:
		return false
	}
}

// AllowedRedirectURIs validates redirect URIs against a configured allow list.
type AllowedRedirectURIs struct {
	uris []string
}

// NewAllowedRedirectURIs creates a validator from a comma-separated URI list.
func NewAllowedRedirectURIs(commaSeparated string) *AllowedRedirectURIs {
	var uris []string
	for _, u := range strings.Split(commaSeparated, ",") {
		u = strings.TrimSpace(u)
		if u != "" {
			uris = append(uris, u)
		}
	}
	return &AllowedRedirectURIs{uris: uris}
}

// IsAllowed checks if the given URI is in the allow list.
func (a *AllowedRedirectURIs) IsAllowed(uri string) bool {
	for _, allowed := range a.uris {
		if uri == allowed {
			return true
		}
	}
	return false
}
