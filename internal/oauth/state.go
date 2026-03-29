// Package oauth provides CSRF state token generation and verification
// for the OAuth provider connect flow.
package oauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// StateMaxAge is the maximum age of a state token before it is considered expired.
const StateMaxAge = 5 * time.Minute

// StateSigner generates and verifies HMAC-signed OAuth state tokens.
type StateSigner struct {
	key []byte
}

// NewStateSigner creates a StateSigner with the given HMAC key.
func NewStateSigner(key []byte) *StateSigner {
	return &StateSigner{key: key}
}

// Sign generates a signed state token embedding the provider ID, tenant ID,
// current timestamp, and a random nonce.
func (s *StateSigner) Sign(providerID, tenantID string) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generating nonce: %w", err)
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonceStr := base64.RawURLEncoding.EncodeToString(nonce)
	payload := providerID + "|" + tenantID + "|" + timestamp + "|" + nonceStr

	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	token := payload + "|" + sig
	return base64.RawURLEncoding.EncodeToString([]byte(token)), nil
}

// Verify decodes and validates a signed state token.
// Returns an error if the signature is invalid, the token is expired,
// or the provider/tenant IDs don't match the expected values.
func (s *StateSigner) Verify(token, expectedProviderID, expectedTenantID string) error {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return fmt.Errorf("invalid state encoding")
	}

	parts := strings.SplitN(string(raw), "|", 5)
	if len(parts) != 5 {
		return fmt.Errorf("invalid state format")
	}

	providerID, tenantID, timestampStr, nonceStr, sig := parts[0], parts[1], parts[2], parts[3], parts[4]

	// Verify signature.
	payload := providerID + "|" + tenantID + "|" + timestampStr + "|" + nonceStr
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(payload))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return fmt.Errorf("invalid state signature")
	}

	// Verify provider ID.
	if providerID != expectedProviderID {
		return fmt.Errorf("state provider mismatch")
	}

	// Verify tenant ID.
	if tenantID != expectedTenantID {
		return fmt.Errorf("state tenant mismatch")
	}

	// Verify timestamp.
	ts, parseErr := strconv.ParseInt(timestampStr, 10, 64)
	if parseErr != nil {
		return fmt.Errorf("invalid state timestamp")
	}
	if time.Since(time.Unix(ts, 0)) > StateMaxAge {
		return fmt.Errorf("state expired")
	}

	return nil
}
