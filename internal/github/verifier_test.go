package github

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
)

// helper untuk buat signature valid
func makeSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifySignature_Valid(t *testing.T) {
	secret := "mysecret"
	payload := []byte(`{"action":"queued"}`)
	signature := makeSignature(secret, payload)

	req, err := http.NewRequest("POST", "/", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("X-Hub-Signature-256", signature)

	if err := VerifySignature(req, secret); err != nil {
		t.Fatalf("expected valid signature, got error: %v", err)
	}
}

func TestVerifySignature_Invalid(t *testing.T) {
	secret := "mysecret"
	payload := []byte(`{"action":"queued"}`)
	// salah secret -> signature mismatch
	badSecret := "wrongsecret"
	signature := makeSignature(badSecret, payload)

	req, err := http.NewRequest("POST", "/", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("X-Hub-Signature-256", signature)

	if err := VerifySignature(req, secret); err == nil {
		t.Fatalf("expected error for invalid signature, got nil")
	}
}

func TestVerifySignature_MissingHeader(t *testing.T) {
	secret := "mysecret"
	payload := []byte(`{"action":"queued"}`)

	req, err := http.NewRequest("POST", "/", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if err := VerifySignature(req, secret); err == nil {
		t.Fatalf("expected error for missing header, got nil")
	}
}
