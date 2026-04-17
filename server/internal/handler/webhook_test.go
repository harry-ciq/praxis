package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifyGitHubSignature(t *testing.T) {
	secret := "topsecret"
	body := []byte(`{"ok":true}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	good := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	cases := []struct {
		name      string
		signature string
		want      bool
	}{
		{"valid", good, true},
		{"missing prefix", hex.EncodeToString(mac.Sum(nil)), false},
		{"empty", "", false},
		{"wrong secret", "sha256=deadbeef", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := verifyGitHubSignature(body, c.signature, secret)
			if got != c.want {
				t.Errorf("verifyGitHubSignature(%q): want %v, got %v", c.signature, c.want, got)
			}
		})
	}
}
