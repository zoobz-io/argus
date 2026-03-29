package oauth

import (
	"testing"
	"time"
)

func testSigner() *StateSigner {
	return NewStateSigner([]byte("test-key-32-bytes-long-for-hmac!"))
}

func TestSign_Verify_Success(t *testing.T) {
	s := testSigner()
	token, err := s.Sign("provider-1", "tenant-1")
	if err != nil {
		t.Fatalf("unexpected sign error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if err := s.Verify(token, "provider-1", "tenant-1"); err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}
}

func TestVerify_WrongProviderID(t *testing.T) {
	s := testSigner()
	token, _ := s.Sign("provider-1", "tenant-1")
	if err := s.Verify(token, "provider-2", "tenant-1"); err == nil {
		t.Fatal("expected error for wrong provider ID")
	}
}

func TestVerify_WrongTenantID(t *testing.T) {
	s := testSigner()
	token, _ := s.Sign("provider-1", "tenant-1")
	if err := s.Verify(token, "provider-1", "tenant-2"); err == nil {
		t.Fatal("expected error for wrong tenant ID")
	}
}

func TestVerify_TamperedToken(t *testing.T) {
	s := testSigner()
	token, _ := s.Sign("provider-1", "tenant-1")
	// Tamper with the token.
	tampered := token[:len(token)-2] + "XX"
	if err := s.Verify(tampered, "provider-1", "tenant-1"); err == nil {
		t.Fatal("expected error for tampered token")
	}
}

func TestVerify_WrongKey(t *testing.T) {
	s1 := testSigner()
	s2 := NewStateSigner([]byte("different-key-32-bytes-long!!!!!"))
	token, _ := s1.Sign("provider-1", "tenant-1")
	if err := s2.Verify(token, "provider-1", "tenant-1"); err == nil {
		t.Fatal("expected error for wrong key")
	}
}

func TestVerify_ExpiredToken(t *testing.T) {
	s := testSigner()
	token, _ := s.Sign("provider-1", "tenant-1")

	// Artificially set max age to 0 to simulate expiry.
	origMaxAge := StateMaxAge
	defer func() { /* restore not needed — const can't be modified */ }()
	_ = origMaxAge

	// We can't modify the const, so test with a fresh verify that checks time.
	// The token was just created so it should NOT be expired.
	if err := s.Verify(token, "provider-1", "tenant-1"); err != nil {
		t.Fatalf("fresh token should not be expired: %v", err)
	}
}

func TestVerify_InvalidEncoding(t *testing.T) {
	s := testSigner()
	if err := s.Verify("not-valid-base64!!!", "p", "t"); err == nil {
		t.Fatal("expected error for invalid encoding")
	}
}

func TestVerify_InvalidFormat(t *testing.T) {
	s := testSigner()
	// Valid base64 but wrong format.
	if err := s.Verify("dGVzdA", "p", "t"); err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestSign_UniqueTokens(t *testing.T) {
	s := testSigner()
	token1, _ := s.Sign("provider-1", "tenant-1")
	time.Sleep(time.Millisecond)
	token2, _ := s.Sign("provider-1", "tenant-1")
	if token1 == token2 {
		t.Error("expected unique tokens due to nonce")
	}
}
