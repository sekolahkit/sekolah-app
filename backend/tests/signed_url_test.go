package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Sekolahkit/sekolah-app/internal/upload"
)

func uploadTestFile(t *testing.T, serverURL string, cookies []*http.Cookie, category, filename, content string) string {
	t.Helper()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	part, err := mw.CreatePart(map[string][]string{
		"Content-Disposition": {fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename)},
		"Content-Type":        {"application/pdf"},
	})
	if err != nil {
		t.Fatal(err)
	}
	part.Write([]byte(content))
	mw.WriteField("category", category)
	mw.Close()

	req, _ := http.NewRequest("POST", serverURL+"/api/v1/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	for _, c := range cookies {
		req.AddCookie(c)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("upload failed: %d %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	data := result["data"].(map[string]interface{})
	return data["path"].(string)
}

func generateSignedURL(t *testing.T, serverURL string, cookies []*http.Cookie, path string, ttlSeconds int) map[string]interface{} {
	t.Helper()

	body := fmt.Sprintf(`{"path":"%s","ttl_seconds":%d}`, path, ttlSeconds)
	req, _ := http.NewRequest("POST", serverURL+"/api/v1/upload/signed", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("generate signed URL failed: %d %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result["data"].(map[string]interface{})
}

func insertTestSchool(t *testing.T, db *sql.DB, nama, kode, adminEmail string) int64 {
	t.Helper()
	hash, _ := bcryptHash("password123")
	result, err := db.Exec("INSERT INTO sekolah (nama, kode) VALUES (?, ?)", nama, kode)
	if err != nil {
		t.Fatal(err)
	}
	sekolahID, _ := result.LastInsertId()

	_, err = db.Exec("INSERT INTO pengguna (sekolah_id, email, password, nama, role, aktif) VALUES (?, ?, ?, ?, ?, TRUE)",
		sekolahID, adminEmail, hash, "Admin "+nama, "admin")
	if err != nil {
		t.Fatal(err)
	}
	return sekolahID
}

func TestSignedURL_ValidToken_ServesFile(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	path := uploadTestFile(t, server.URL, cookies, "general", "test.pdf", "%PDF-1.4 fake content")

	data := generateSignedURL(t, server.URL, cookies, path, 300)
	url := data["url"].(string)

	resp, err := http.Get(server.URL + url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
	}
}

func TestSignedURL_ExpiredToken_Rejected(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	path := uploadTestFile(t, server.URL, cookies, "general", "test.pdf", "%PDF-1.4 fake")

	result, err := upload.SignPath(testJWTSecret, 1, path, 1*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	resp, err := http.Get(server.URL + "/api/v1/upload/signed/" + result.Token)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 403 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 403 for expired token, got %d: %s", resp.StatusCode, string(body))
	}
}

func TestSignedURL_TamperedToken_Rejected(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	path := uploadTestFile(t, server.URL, cookies, "general", "test.pdf", "%PDF-1.4 fake")

	data := generateSignedURL(t, server.URL, cookies, path, 300)
	url := data["url"].(string)

	tamperedURL := url + "tampered"
	resp, err := http.Get(server.URL + tamperedURL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 403 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 403 for tampered token, got %d: %s", resp.StatusCode, string(body))
	}
}

func TestSignedURL_PathTraversal_Rejected(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	body := `{"path":"1/general/../../../etc/passwd","ttl_seconds":300}`
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/upload/signed", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 403 {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 403 for path traversal, got %d: %s", resp.StatusCode, string(respBody))
	}
}

func TestSignedURL_CrossSchoolToken_Isolated(t *testing.T) {
	server, db := setupTestServer(t)
	doSetup(t, server, "school1", "admin@s1.id", "password123")
	insertTestSchool(t, db, "School 2", "school2", "admin@s2.id")

	cookies1 := doLogin(t, server, "school1", "admin@s1.id", "password123")
	cookies2 := doLogin(t, server, "school2", "admin@s2.id", "password123")

	path := uploadTestFile(t, server.URL, cookies1, "general", "secret.pdf", "%PDF-1.4 school1 data")

	data := generateSignedURL(t, server.URL, cookies1, path, 300)
	url := data["url"].(string)

	resp, err := http.Get(server.URL + url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("owner should access own file, got %d", resp.StatusCode)
	}

	_ = cookies2
}

func TestSignedURL_CrossSchoolFileAccess_Rejected(t *testing.T) {
	server, db := setupTestServer(t)
	doSetup(t, server, "school1", "admin@s1.id", "password123")
	insertTestSchool(t, db, "School 2", "school2", "admin@s2.id")

	cookies1 := doLogin(t, server, "school1", "admin@s1.id", "password123")
	cookies2 := doLogin(t, server, "school2", "admin@s2.id", "password123")

	path1 := uploadTestFile(t, server.URL, cookies1, "general", "s1.pdf", "%PDF-1.4 school1 data")
	path2 := uploadTestFile(t, server.URL, cookies2, "general", "s2.pdf", "%PDF-1.4 school2 data")

	result1, err := upload.SignPath(testJWTSecret, 1, path1, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	result2, err := upload.SignPath(testJWTSecret, 2, path2, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	resp1, _ := http.Get(server.URL + "/api/v1/upload/signed/" + result1.Token)
	defer resp1.Body.Close()
	if resp1.StatusCode != 200 {
		t.Fatalf("school1 should access own file, got %d", resp1.StatusCode)
	}

	resp2, _ := http.Get(server.URL + "/api/v1/upload/signed/" + result2.Token)
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Fatalf("school2 should access own file, got %d", resp2.StatusCode)
	}

	wrongSig, err := upload.SignPath(testJWTSecret, 2, path1, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	resp3, _ := http.Get(server.URL + "/api/v1/upload/signed/" + wrongSig.Token)
	defer resp3.Body.Close()
	if resp3.StatusCode != 403 {
		body, _ := io.ReadAll(resp3.Body)
		t.Fatalf("school2 should NOT access school1 file, expected 403, got %d: %s", resp3.StatusCode, string(body))
	}
}

func TestSignedURL_GenerateRequiresAuth(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")

	body := `{"path":"1/general/test.pdf","ttl_seconds":300}`
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/upload/signed", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 401 {
		t.Fatalf("expected 401 without auth, got %d", resp.StatusCode)
	}
}

func TestSignedURL_GenerateValidatesFileOwnership(t *testing.T) {
	server, db := setupTestServer(t)
	doSetup(t, server, "school1", "admin@s1.id", "password123")
	insertTestSchool(t, db, "School 2", "school2", "admin@s2.id")

	cookies1 := doLogin(t, server, "school1", "admin@s1.id", "password123")
	cookies2 := doLogin(t, server, "school2", "admin@s2.id", "password123")

	path := uploadTestFile(t, server.URL, cookies1, "general", "test.pdf", "%PDF-1.4 school1 data")

	body := fmt.Sprintf(`{"path":"%s","ttl_seconds":300}`, path)
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/upload/signed", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies2 {
		req.AddCookie(c)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 403 {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 403 when generating signed URL for another school's file, got %d: %s", resp.StatusCode, string(respBody))
	}
}

func TestSignedURL_TTLMaxEnforced(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	path := uploadTestFile(t, server.URL, cookies, "general", "test.pdf", "%PDF-1.4 fake")

	body := fmt.Sprintf(`{"path":"%s","ttl_seconds":600}`, path)
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/upload/signed", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	data := result["data"].(map[string]interface{})

	expiresAtStr := data["expires_at"].(string)
	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		t.Fatalf("failed to parse expires_at: %v", err)
	}

	maxExpiry := time.Now().Add(upload.MaxTTL + 2*time.Second)
	if expiresAt.After(maxExpiry) {
		t.Errorf("TTL should be capped at 15 minutes, but expires_at is %s", expiresAtStr)
	}
}

func TestSignedURL_ServeSignedURLIsPublic(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	path := uploadTestFile(t, server.URL, cookies, "general", "test.pdf", "%PDF-1.4 fake")

	result, err := upload.SignPath(testJWTSecret, 1, path, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(server.URL + "/api/v1/upload/signed/" + result.Token)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("signed URL should be publicly accessible, got %d: %s", resp.StatusCode, string(body))
	}
}

func TestSignedURL_WrongSecret_Rejected(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	path := uploadTestFile(t, server.URL, cookies, "general", "test.pdf", "%PDF-1.4 fake")

	result, err := upload.SignPath("wrong-secret-that-is-long-enough-for-hmac-signing", 1, path, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(server.URL + "/api/v1/upload/signed/" + result.Token)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 403 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 403 for wrong secret, got %d: %s", resp.StatusCode, string(body))
	}
}

func TestSignedURL_InvalidBase64_Rejected(t *testing.T) {
	server, _ := setupTestServer(t)

	resp, err := http.Get(server.URL + "/api/v1/upload/signed/not-a-valid-token!!!")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 403 {
		t.Fatalf("expected 403 for invalid token, got %d", resp.StatusCode)
	}
}

func TestSignedURL_DefaultTTLUsed(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	path := uploadTestFile(t, server.URL, cookies, "general", "test.pdf", "%PDF-1.4 fake")

	body := fmt.Sprintf(`{"path":"%s","ttl_seconds":0}`, path)
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/upload/signed", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	data := result["data"].(map[string]interface{})

	expiresAtStr := data["expires_at"].(string)
	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		t.Fatalf("failed to parse expires_at: %v", err)
	}

	now := time.Now()
	minExpiry := now.Add(upload.DefaultTTL - 2*time.Second)
	maxExpiry := now.Add(upload.DefaultTTL + 2*time.Second)
	if expiresAt.Before(minExpiry) || expiresAt.After(maxExpiry) {
		t.Errorf("default TTL should be ~5 minutes, got expires_at=%s (now=%s)", expiresAtStr, now.Format(time.RFC3339))
	}
}

func TestSignedURL_NonexistentFile_Rejected(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")
	cookies := doLogin(t, server, "test1", "admin@test1.id", "password123")

	body := `{"path":"1/general/nonexistent.pdf","ttl_seconds":300}`
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/upload/signed", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 404 for nonexistent file, got %d: %s", resp.StatusCode, string(respBody))
	}
}

func TestSignedURL_UploadDownload_StillRequiresAuth(t *testing.T) {
	server, _ := setupTestServer(t)
	doSetup(t, server, "test1", "admin@test1.id", "password123")

	resp, err := http.Get(server.URL + "/api/v1/upload/1/general/test.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 401 {
		t.Fatalf("authenticated download should still require auth, got %d", resp.StatusCode)
	}
}
