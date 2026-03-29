package wire

import (
	"strings"
	"testing"
)

func TestAuthURLResponse_Clone(t *testing.T) {
	orig := AuthURLResponse{URL: "https://accounts.google.com/o/oauth2/auth?client_id=123"}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestConnectRequest_Validate_Valid(t *testing.T) {
	req := ConnectRequest{
		Code:        "auth-code-123",
		RedirectURI: "https://app.example.com/callback",
		State:       "csrf-token",
	}
	if err := req.Validate(); err != nil {
		t.Errorf("expected valid, got error: %v", err)
	}
}

func TestConnectRequest_Validate_MissingCode(t *testing.T) {
	req := ConnectRequest{
		RedirectURI: "https://app.example.com/callback",
		State:       "csrf-token",
	}
	err := req.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "code") {
		t.Errorf("expected error to mention %q, got: %v", "code", err)
	}
}

func TestConnectRequest_Validate_MissingState(t *testing.T) {
	req := ConnectRequest{
		Code:        "auth-code-123",
		RedirectURI: "https://app.example.com/callback",
	}
	err := req.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "state") {
		t.Errorf("expected error to mention %q, got: %v", "state", err)
	}
}

func TestConnectRequest_Clone(t *testing.T) {
	orig := ConnectRequest{Code: "code", RedirectURI: "https://example.com/cb", State: "state"}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestConnectCredentialsRequest_Validate_Valid(t *testing.T) {
	req := ConnectCredentialsRequest{
		Credentials: `{"access_key":"ABC","secret_key":"XYZ"}`,
	}
	if err := req.Validate(); err != nil {
		t.Errorf("expected valid, got error: %v", err)
	}
}

func TestConnectCredentialsRequest_Validate_MissingCredentials(t *testing.T) {
	req := ConnectCredentialsRequest{}
	err := req.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "credentials") {
		t.Errorf("expected error to mention %q, got: %v", "credentials", err)
	}
}

func TestConnectCredentialsRequest_Clone(t *testing.T) {
	orig := ConnectCredentialsRequest{Credentials: `{"key":"value"}`}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestConnectResponse_Clone(t *testing.T) {
	orig := ConnectResponse{Status: "connected"}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}
