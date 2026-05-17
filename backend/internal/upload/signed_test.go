package upload

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestSignPath_ValidSignature(t *testing.T) {
	result, err := SignPath("test-secret-key-minimum-32-chars", 1, "1/general/test.txt", 5*time.Minute)
	if err != nil {
		t.Fatalf("SignPath failed: %v", err)
	}

	if result.Token == "" {
		t.Error("token should not be empty")
	}
	if !strings.HasPrefix(result.URL, "/api/v1/upload/signed/") {
		t.Errorf("URL should start with /api/v1/upload/signed/, got %s", result.URL)
	}
	if result.ExpiresAt == "" {
		t.Error("expires_at should not be empty")
	}
}

func TestSignPath_TTLClamped(t *testing.T) {
	result, err := SignPath("test-secret-key-minimum-32-chars", 1, "1/general/test.txt", 1*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	expiresAt, _ := time.Parse(time.RFC3339, result.ExpiresAt)
	maxExpiry := time.Now().Add(MaxTTL + 2*time.Second)
	if expiresAt.After(maxExpiry) {
		t.Errorf("TTL should be clamped to MaxTTL, got %s", result.ExpiresAt)
	}
}

func TestSignPath_DefaultTTL(t *testing.T) {
	result, err := SignPath("test-secret-key-minimum-32-chars", 1, "1/general/test.txt", 0)
	if err != nil {
		t.Fatal(err)
	}

	expiresAt, _ := time.Parse(time.RFC3339, result.ExpiresAt)
	now := time.Now()
	minExpiry := now.Add(DefaultTTL - 2*time.Second)
	maxExpiry := now.Add(DefaultTTL + 2*time.Second)
	if expiresAt.Before(minExpiry) || expiresAt.After(maxExpiry) {
		t.Errorf("default TTL should be ~5 minutes, got %s", result.ExpiresAt)
	}
}

func TestValidateSignedToken_Valid(t *testing.T) {
	secret := "test-secret-key-minimum-32-chars"
	result, _ := SignPath(secret, 1, "1/general/test.txt", 5*time.Minute)

	payload, err := ValidateSignedToken(secret, result.Token)
	if err != nil {
		t.Fatalf("ValidateSignedToken failed: %v", err)
	}
	if payload.SekolahID != 1 {
		t.Errorf("expected sekolah_id=1, got %d", payload.SekolahID)
	}
	if payload.Path != "1/general/test.txt" {
		t.Errorf("expected path '1/general/test.txt', got '%s'", payload.Path)
	}
}

func TestValidateSignedToken_Expired(t *testing.T) {
	secret := "test-secret-key-minimum-32-chars"
	result, _ := SignPath(secret, 1, "1/general/test.txt", 1*time.Second)

	time.Sleep(2 * time.Second)

	_, err := ValidateSignedToken(secret, result.Token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	if !strings.Contains(err.Error(), "kedaluwarsa") {
		t.Errorf("expected 'kedaluwarsa' in error, got: %v", err)
	}
}

func TestValidateSignedToken_WrongSecret(t *testing.T) {
	result, _ := SignPath("correct-secret-key-minimum-32", 1, "1/general/test.txt", 5*time.Minute)

	_, err := ValidateSignedToken("wrong-secret-key-minimum-32-ch", result.Token)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
	if !strings.Contains(err.Error(), "signature") {
		t.Errorf("expected signature error, got: %v", err)
	}
}

func TestValidateSignedToken_InvalidFormat(t *testing.T) {
	secret := "test-secret-key-minimum-32-chars"

	_, err := ValidateSignedToken(secret, "no-dot-separator")
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestValidateSignedToken_InvalidBase64(t *testing.T) {
	secret := "test-secret-key-minimum-32-chars"

	_, err := ValidateSignedToken(secret, "!!!.!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestValidateSignedToken_PathTraversal(t *testing.T) {
	secret := "test-secret-key-minimum-32-chars"

	_, err := SignPath(secret, 1, "1/general/../../../etc/passwd", 5*time.Minute)
	if err != nil {
		t.Fatalf("SignPath should not error, validation happens at token use: %v", err)
	}
}

func TestValidatePath_Valid(t *testing.T) {
	err := ValidatePath("1/general/test.txt")
	if err != nil {
		t.Errorf("expected no error for valid path, got: %v", err)
	}
}

func TestValidatePath_Empty(t *testing.T) {
	err := ValidatePath("")
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestValidatePath_Traversal(t *testing.T) {
	err := ValidatePath("1/general/../../../etc/passwd")
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestValidatePath_NoSlash(t *testing.T) {
	err := ValidatePath("noseparator")
	if err == nil {
		t.Error("expected error for path without separator")
	}
}

func TestValidateSignedToken_NegativeSekolahID(t *testing.T) {
	secret := "test-secret-key-minimum-32-chars"

	payload := signedPayload{
		SekolahID: -1,
		Path:      "1/general/test.txt",
		ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
	}

	payloadJSON, _ := json.Marshal(payload)
	sig := computeHMAC(secret, payloadJSON)
	token := b64.EncodeToString(payloadJSON) + "." + b64.EncodeToString(sig)

	_, err := ValidateSignedToken(secret, token)
	if err == nil {
		t.Fatal("expected error for negative sekolah_id")
	}
}
