package wire

import (
	"net/url"
	"strings"

	"github.com/zoobz-io/cereal"
	"github.com/zoobz-io/sum"
)

// Custom mask types for the webhook wire types.
const (
	MaskURL    cereal.MaskType = "url"
	MaskSecret cereal.MaskType = "secret"
)

// redacted is the default replacement for masked values.
const redacted = "***"

// urlMasker masks URLs: https://example.com/path -> https://e***/***
type urlMasker struct{}

// URLMasker returns a masker for URLs.
func URLMasker() cereal.Masker {
	return &urlMasker{}
}

func (m *urlMasker) Mask(value string) (string, error) {
	u, err := url.Parse(value)
	if err != nil || u.Host == "" {
		return redacted, nil
	}
	host := u.Hostname()
	masked := string(host[0]) + redacted
	if port := u.Port(); port != "" {
		masked += ":" + port
	}
	return u.Scheme + "://" + masked + "/" + redacted, nil
}

// secretMasker fully redacts secret values.
type secretMasker struct{}

// SecretMasker returns a masker that fully redacts a value.
func SecretMasker() cereal.Masker {
	return &secretMasker{}
}

func (m *secretMasker) Mask(value string) (string, error) {
	if len(value) == 0 {
		return redacted, nil
	}
	return strings.Repeat("*", len(value)), nil
}

// RegisterMaskers registers the custom mask types on the sum service.
func RegisterMaskers(svc *sum.Service) {
	svc.WithMasker(MaskURL, URLMasker())
	svc.WithMasker(MaskSecret, SecretMasker())
}
