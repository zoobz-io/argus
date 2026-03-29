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

// urlMasker masks URLs: https://example.com/path -> https://e***.com/***
type urlMasker struct{}

// URLMasker returns a masker for URLs.
func URLMasker() cereal.Masker {
	return &urlMasker{}
}

func (m *urlMasker) Mask(value string) (string, error) {
	u, err := url.Parse(value)
	if err != nil || u.Host == "" {
		return "***", nil
	}
	host := u.Hostname()
	masked := string(host[0]) + "***"
	if port := u.Port(); port != "" {
		masked += ":" + port
	}
	return u.Scheme + "://" + masked + "/***", nil
}

// secretMasker fully redacts secret values.
type secretMasker struct{}

// SecretMasker returns a masker that fully redacts a value.
func SecretMasker() cereal.Masker {
	return &secretMasker{}
}

func (m *secretMasker) Mask(value string) (string, error) {
	if len(value) == 0 {
		return "***", nil
	}
	return strings.Repeat("*", len(value)), nil
}

// RegisterMaskers registers the custom mask types on the sum service.
func RegisterMaskers(svc *sum.Service) {
	svc.WithMasker(MaskURL, URLMasker())
	svc.WithMasker(MaskSecret, SecretMasker())
}
