package config

// Providers holds OAuth app credentials for cloud storage providers.
// These are per-application, not per-tenant. One set per provider type.
type Providers struct {
	GoogleClientID        string `env:"APP_GOOGLE_CLIENT_ID"`
	GoogleClientSecret    string `env:"APP_GOOGLE_CLIENT_SECRET"`
	MicrosoftClientID     string `env:"APP_MICROSOFT_CLIENT_ID"`
	MicrosoftClientSecret string `env:"APP_MICROSOFT_CLIENT_SECRET"`
	DropboxClientID       string `env:"APP_DROPBOX_CLIENT_ID"`
	DropboxClientSecret   string `env:"APP_DROPBOX_CLIENT_SECRET"`
	AllowedRedirectURIs   string `env:"APP_OAUTH_REDIRECT_URIS" default:"http://localhost:3000/providers/callback"`
}

// Validate checks that the configuration is valid.
// All fields are optional — only needed if the corresponding provider is used.
func (c Providers) Validate() error {
	return nil
}
