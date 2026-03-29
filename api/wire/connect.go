package wire

import "github.com/zoobz-io/check"

// AuthURLResponse is the response for the OAuth authorization URL endpoint.
type AuthURLResponse struct {
	URL string `json:"url" description:"OAuth authorization URL to redirect the user to"`
}

// Clone returns a copy of the response.
func (r AuthURLResponse) Clone() AuthURLResponse { return r }

// ConnectRequest is the request body for connecting an OAuth provider.
type ConnectRequest struct {
	Code        string `json:"code" description:"Authorization code from OAuth callback"`
	RedirectURI string `json:"redirect_uri" description:"Redirect URI used in the OAuth flow"`
	State       string `json:"state" description:"CSRF state token from the OAuth callback"`
}

// Validate validates the request fields.
func (r *ConnectRequest) Validate() error {
	return check.All(
		check.Str(r.Code, "code").Required().V(),
		check.Str(r.RedirectURI, "redirect_uri").Required().V(),
		check.Str(r.State, "state").Required().V(),
	).Err()
}

// Clone returns a copy of the request.
func (r ConnectRequest) Clone() ConnectRequest { return r }

// ConnectCredentialsRequest is the request body for connecting a static credential provider.
type ConnectCredentialsRequest struct {
	Credentials string `json:"credentials" description:"Provider credentials (JSON)"`
}

// Validate validates the request fields.
func (r *ConnectCredentialsRequest) Validate() error {
	return check.All(
		check.Str(r.Credentials, "credentials").Required().V(),
	).Err()
}

// Clone returns a copy of the request.
func (r ConnectCredentialsRequest) Clone() ConnectCredentialsRequest { return r }

// ConnectResponse is the response for a successful provider connection.
type ConnectResponse struct {
	Status string `json:"status" description:"Connection status"`
}

// Clone returns a copy of the response.
func (r ConnectResponse) Clone() ConnectResponse { return r }
