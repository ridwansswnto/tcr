package github

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
)

func VerifySignature(r *http.Request, secret string) error {
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		return errors.New("missing signature header")
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return errors.New("invalid signature")
	}

	r.Body = io.NopCloser(bytes.NewReader(payload))
	return nil
}
