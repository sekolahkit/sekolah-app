package upload

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	DefaultTTL = 5 * time.Minute
	MaxTTL     = 15 * time.Minute
)

type signedPayload struct {
	SekolahID int64  `json:"sid"`
	Path      string `json:"p"`
	ExpiresAt int64  `json:"exp"`
}

type SignedResult struct {
	Token     string `json:"token"`
	URL       string `json:"url"`
	ExpiresAt string `json:"expires_at"`
}

var b64 = base64.URLEncoding.WithPadding(base64.NoPadding)

func SignPath(secret string, sekolahID int64, path string, ttl time.Duration) (*SignedResult, error) {
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	if ttl > MaxTTL {
		ttl = MaxTTL
	}

	payload := signedPayload{
		SekolahID: sekolahID,
		Path:      path,
		ExpiresAt: time.Now().Add(ttl).Unix(),
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	sig := computeHMAC(secret, payloadJSON)
	token := b64.EncodeToString(payloadJSON) + "." + b64.EncodeToString(sig)

	return &SignedResult{
		Token:     token,
		URL:       "/api/v1/upload/signed/" + token,
		ExpiresAt: time.Unix(payload.ExpiresAt, 0).UTC().Format(time.RFC3339),
	}, nil
}

func ValidateSignedToken(secret, token string) (*signedPayload, error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("token format tidak valid")
	}

	payloadBytes, err := b64.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("token payload tidak valid")
	}

	sigBytes, err := b64.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("token signature tidak valid")
	}

	expectedSig := computeHMAC(secret, payloadBytes)
	if !hmac.Equal(sigBytes, expectedSig) {
		return nil, fmt.Errorf("token signature tidak cocok")
	}

	var payload signedPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("token payload tidak valid")
	}

	if time.Now().Unix() > payload.ExpiresAt {
		return nil, fmt.Errorf("token sudah kedaluwarsa")
	}

	if payload.SekolahID <= 0 {
		return nil, fmt.Errorf("token tidak valid")
	}

	if err := ValidatePath(payload.Path); err != nil {
		return nil, err
	}

	pathParts := strings.SplitN(payload.Path, "/", 2)
	if pathParts[0] != fmt.Sprintf("%d", payload.SekolahID) {
		return nil, fmt.Errorf("token tidak valid: sekolah_id tidak sesuai")
	}

	return &payload, nil
}

func ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path tidak boleh kosong")
	}
	if strings.Contains(path, "..") {
		return fmt.Errorf("path tidak valid")
	}
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return fmt.Errorf("path tidak valid")
	}
	return nil
}

func computeHMAC(secret string, data []byte) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(data)
	return mac.Sum(nil)
}
