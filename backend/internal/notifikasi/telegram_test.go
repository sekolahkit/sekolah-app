package notifikasi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTelegramTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE notifikasi_preferensi (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			pengguna_id INTEGER,
			siswa_id INTEGER,
			recipient_type TEXT NOT NULL DEFAULT 'manual',
			channel TEXT NOT NULL,
			destination TEXT NOT NULL,
			enabled BOOLEAN NOT NULL DEFAULT TRUE,
			consent_status TEXT NOT NULL DEFAULT 'pending',
			consent_source TEXT NOT NULL DEFAULT 'admin',
			consent_at DATETIME,
			revoked_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE UNIQUE INDEX idx_pref_unique ON notifikasi_preferensi(sekolah_id, channel, destination);

		CREATE TABLE telegram_invite (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			preference_id INTEGER NOT NULL,
			token_hash TEXT NOT NULL UNIQUE,
			invite_expires_at DATETIME NOT NULL,
			invite_used_at DATETIME,
			telegram_user_id INTEGER,
			telegram_chat_id INTEGER,
			telegram_username TEXT,
			status TEXT NOT NULL DEFAULT 'pending_invite',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_tg_invite_hash ON telegram_invite(token_hash);
		CREATE INDEX idx_tg_invite_chat ON telegram_invite(telegram_chat_id);
	`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func insertPreference(t *testing.T, db *sql.DB, sekolahID int64, channel, destination, consentStatus string) int64 {
	t.Helper()
	result, err := db.Exec(`INSERT INTO notifikasi_preferensi (sekolah_id, channel, destination, enabled, consent_status, consent_source) VALUES (?, ?, ?, 1, ?, 'admin')`,
		sekolahID, channel, destination, consentStatus)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := result.LastInsertId()
	return id
}

func TestGenerateInviteToken(t *testing.T) {
	token, err := generateToken()
	if err != nil {
		t.Fatalf("generateToken failed: %v", err)
	}
	if len(token) != 48 {
		t.Fatalf("expected 48 hex chars, got %d", len(token))
	}
}

func TestHashToken(t *testing.T) {
	token := "test-token-123"
	hash1 := hashToken(token)
	hash2 := hashToken(token)

	if hash1 != hash2 {
		t.Fatal("same token should produce same hash")
	}
	if len(hash1) != 64 {
		t.Fatalf("expected 64 hex chars (SHA256), got %d", len(hash1))
	}
	if hash1 == token {
		t.Fatal("hash should not equal original token")
	}
}

func TestDifferentTokensDifferentHashes(t *testing.T) {
	hash1 := hashToken("token-a")
	hash2 := hashToken("token-b")
	if hash1 == hash2 {
		t.Fatal("different tokens should produce different hashes")
	}
}

func TestGenerateInviteLink(t *testing.T) {
	db := setupTelegramTestDB(t)
	defer db.Close()

	inviteRepo := NewTelegramInviteRepository(db)
	prefRepo := NewPreferensiRepository(db)
	prefID := insertPreference(t, db, 1, "telegram", "pending", "pending")

	cfg := TelegramConfig{BotToken: "test-token", BotUsername: "TestBot"}
	svc := NewTelegramService(inviteRepo, prefRepo, cfg)

	link, err := svc.GenerateInviteLink(1, prefID)
	if err != nil {
		t.Fatalf("GenerateInviteLink failed: %v", err)
	}

	if link == "" {
		t.Fatal("expected non-empty link")
	}
	if link[:len("https://t.me/TestBot?start=")] != "https://t.me/TestBot?start=" {
		t.Fatalf("unexpected link format: %s", link)
	}
}

func TestInviteTokenIsHashedAtRest(t *testing.T) {
	db := setupTelegramTestDB(t)
	defer db.Close()

	inviteRepo := NewTelegramInviteRepository(db)
	prefRepo := NewPreferensiRepository(db)
	prefID := insertPreference(t, db, 1, "telegram", "pending", "pending")

	cfg := TelegramConfig{BotToken: "t", BotUsername: "Bot"}
	svc := NewTelegramService(inviteRepo, prefRepo, cfg)

	link, _ := svc.GenerateInviteLink(1, prefID)
	token := link[len("https://t.me/Bot?start="):]

	var storedHash string
	err := db.QueryRow(`SELECT token_hash FROM telegram_invite WHERE preference_id = ?`, prefID).Scan(&storedHash)
	if err != nil {
		t.Fatal(err)
	}

	if storedHash == token {
		t.Fatal("token should be stored as hash, not plaintext")
	}
	if storedHash != hashToken(token) {
		t.Fatal("stored hash should match SHA256 of token")
	}
}

func TestInviteTokenIsSingleUse(t *testing.T) {
	db := setupTelegramTestDB(t)
	defer db.Close()

	inviteRepo := NewTelegramInviteRepository(db)
	prefRepo := NewPreferensiRepository(db)
	prefID := insertPreference(t, db, 1, "telegram", "pending", "pending")

	cfg := TelegramConfig{BotToken: "t", BotUsername: "Bot"}
	svc := NewTelegramService(inviteRepo, prefRepo, cfg)

	link, _ := svc.GenerateInviteLink(1, prefID)
	token := link[len("https://t.me/Bot?start="):]
	hash := hashToken(token)

	inv, err := inviteRepo.GetByTokenHash(hash)
	if err != nil {
		t.Fatalf("GetByTokenHash failed: %v", err)
	}
	if inv.Status != telegramStatusPending {
		t.Fatalf("expected pending_invite, got %s", inv.Status)
	}

	err = inviteRepo.MarkUsed(inv.ID, 100, 200, "testuser")
	if err != nil {
		t.Fatalf("MarkUsed failed: %v", err)
	}

	inv2, _ := inviteRepo.GetByTokenHash(hash)
	if inv2.Status != telegramStatusGranted {
		t.Fatalf("expected granted after use, got %s", inv2.Status)
	}
	if inv2.InviteUsedAt == "" {
		t.Fatal("expected invite_used_at to be set")
	}
	if inv2.TelegramUserID != 100 {
		t.Fatalf("expected telegram_user_id 100, got %d", inv2.TelegramUserID)
	}
	if inv2.TelegramChatID != 200 {
		t.Fatalf("expected telegram_chat_id 200, got %d", inv2.TelegramChatID)
	}
}

func TestWebhookStartValidToken(t *testing.T) {
	db := setupTelegramTestDB(t)
	defer db.Close()

	inviteRepo := NewTelegramInviteRepository(db)
	prefRepo := NewPreferensiRepository(db)
	prefID := insertPreference(t, db, 1, "telegram", "pending", "pending")

	cfg := TelegramConfig{BotToken: "t", BotUsername: "Bot"}
	svc := NewTelegramService(inviteRepo, prefRepo, cfg)

	link, _ := svc.GenerateInviteLink(1, prefID)
	token := link[len("https://t.me/Bot?start="):]

	var sentTo int64
	var sentText string
	svc.sendFunc = func(_ string, chatID int64, text string) error {
		sentTo = chatID
		sentText = text
		return nil
	}

	update := map[string]interface{}{
		"message": map[string]interface{}{
			"chat": map[string]interface{}{"id": float64(555), "type": "private"},
			"from": map[string]interface{}{"id": float64(111), "username": "parent1"},
			"text":  "/start " + token,
		},
	}
	body, _ := json.Marshal(update)

	err := svc.HandleWebhook(body, "")
	if err != nil {
		t.Fatalf("HandleWebhook failed: %v", err)
	}

	if sentTo != 555 {
		t.Fatalf("expected reply to chat 555, got %d", sentTo)
	}
	if sentText == "" {
		t.Fatal("expected a reply message")
	}

	inv, _ := inviteRepo.GetByTokenHash(hashToken(token))
	if inv.Status != telegramStatusGranted {
		t.Fatalf("expected granted, got %s", inv.Status)
	}
	if inv.TelegramChatID != 555 {
		t.Fatalf("expected chat_id 555, got %d", inv.TelegramChatID)
	}

	var prefDest string
	var prefConsent string
	err = db.QueryRow(`SELECT destination, consent_status FROM notifikasi_preferensi WHERE id = ?`, prefID).Scan(&prefDest, &prefConsent)
	if err != nil {
		t.Fatal(err)
	}
	if prefDest != "555" {
		t.Fatalf("expected destination 555, got %s", prefDest)
	}
	if prefConsent != "granted" {
		t.Fatalf("expected consent granted, got %s", prefConsent)
	}
}

func TestWebhookInvalidToken(t *testing.T) {
	db := setupTelegramTestDB(t)
	defer db.Close()

	inviteRepo := NewTelegramInviteRepository(db)
	prefRepo := NewPreferensiRepository(db)

	cfg := TelegramConfig{BotToken: "t", BotUsername: "Bot"}
	svc := NewTelegramService(inviteRepo, prefRepo, cfg)

	var sentText string
	svc.sendFunc = func(_ string, _ int64, text string) error {
		sentText = text
		return nil
	}

	update := map[string]interface{}{
		"message": map[string]interface{}{
			"chat": map[string]interface{}{"id": float64(555), "type": "private"},
			"from": map[string]interface{}{"id": float64(111), "username": "u"},
			"text":  "/start invalid-token-hash",
		},
	}
	body, _ := json.Marshal(update)

	err := svc.HandleWebhook(body, "")
	if err != nil {
		t.Fatalf("HandleWebhook should not return error for invalid token: %v", err)
	}

	if sentText == "" {
		t.Fatal("expected error reply to user")
	}
}

func TestWebhookExpiredToken(t *testing.T) {
	db := setupTelegramTestDB(t)
	defer db.Close()

	inviteRepo := NewTelegramInviteRepository(db)
	prefRepo := NewPreferensiRepository(db)
	prefID := insertPreference(t, db, 1, "telegram", "pending", "pending")

	token, _ := generateToken()
	hash := hashToken(token)
	expired := time.Now().UTC().Add(-1 * time.Hour).Format("2006-01-02 15:04:05")

	_, err := db.Exec(`INSERT INTO telegram_invite (sekolah_id, preference_id, token_hash, invite_expires_at, status) VALUES (?, ?, ?, ?, 'pending_invite')`,
		1, prefID, hash, expired)
	if err != nil {
		t.Fatal(err)
	}

	cfg := TelegramConfig{BotToken: "t", BotUsername: "Bot"}
	svc := NewTelegramService(inviteRepo, prefRepo, cfg)

	var sentText string
	svc.sendFunc = func(_ string, _ int64, text string) error {
		sentText = text
		return nil
	}

	update := map[string]interface{}{
		"message": map[string]interface{}{
			"chat": map[string]interface{}{"id": float64(555), "type": "private"},
			"from": map[string]interface{}{"id": float64(111), "username": "u"},
			"text":  "/start " + token,
		},
	}
	body, _ := json.Marshal(update)

	svc.HandleWebhook(body, "")

	if sentText == "" {
		t.Fatal("expected reply about expired token")
	}
}

func TestWebhookAlreadyUsedToken(t *testing.T) {
	db := setupTelegramTestDB(t)
	defer db.Close()

	inviteRepo := NewTelegramInviteRepository(db)
	prefRepo := NewPreferensiRepository(db)
	prefID := insertPreference(t, db, 1, "telegram", "pending", "pending")

	token, _ := generateToken()
	hash := hashToken(token)
	expires := time.Now().UTC().Add(1 * time.Hour).Format("2006-01-02 15:04:05")

	result, _ := db.Exec(`INSERT INTO telegram_invite (sekolah_id, preference_id, token_hash, invite_expires_at, status) VALUES (?, ?, ?, ?, 'granted')`,
		1, prefID, hash, expires)
	invID, _ := result.LastInsertId()

	db.Exec(`UPDATE telegram_invite SET invite_used_at = CURRENT_TIMESTAMP, telegram_user_id = 99, telegram_chat_id = 88 WHERE id = ?`, invID)

	cfg := TelegramConfig{BotToken: "t", BotUsername: "Bot"}
	svc := NewTelegramService(inviteRepo, prefRepo, cfg)

	var sentText string
	svc.sendFunc = func(_ string, _ int64, text string) error {
		sentText = text
		return nil
	}

	update := map[string]interface{}{
		"message": map[string]interface{}{
			"chat": map[string]interface{}{"id": float64(555), "type": "private"},
			"from": map[string]interface{}{"id": float64(111), "username": "u"},
			"text":  "/start " + token,
		},
	}
	body, _ := json.Marshal(update)

	svc.HandleWebhook(body, "")

	if sentText == "" {
		t.Fatal("expected reply about already used token")
	}
}

func TestWebhookStopRevokesConsent(t *testing.T) {
	db := setupTelegramTestDB(t)
	defer db.Close()

	inviteRepo := NewTelegramInviteRepository(db)
	prefRepo := NewPreferensiRepository(db)
	prefID := insertPreference(t, db, 1, "telegram", "999", "granted")

	token, _ := generateToken()
	hash := hashToken(token)
	expires := time.Now().UTC().Add(1 * time.Hour).Format("2006-01-02 15:04:05")
	db.Exec(`INSERT INTO telegram_invite (sekolah_id, preference_id, token_hash, invite_expires_at, telegram_chat_id, status) VALUES (?, ?, ?, ?, 999, 'granted')`,
		1, prefID, hash, expires)

	cfg := TelegramConfig{BotToken: "t", BotUsername: "Bot"}
	svc := NewTelegramService(inviteRepo, prefRepo, cfg)

	var sentText string
	svc.sendFunc = func(_ string, _ int64, text string) error {
		sentText = text
		return nil
	}

	update := map[string]interface{}{
		"message": map[string]interface{}{
			"chat": map[string]interface{}{"id": float64(999), "type": "private"},
			"from": map[string]interface{}{"id": float64(111), "username": "u"},
			"text":  "/stop",
		},
	}
	body, _ := json.Marshal(update)

	svc.HandleWebhook(body, "")

	if sentText == "" {
		t.Fatal("expected reply about stopping")
	}

	var consent string
	db.QueryRow(`SELECT consent_status FROM notifikasi_preferensi WHERE id = ?`, prefID).Scan(&consent)
	if consent != "revoked" {
		t.Fatalf("expected consent revoked, got %s", consent)
	}
}

func TestWebhookStatusLinked(t *testing.T) {
	db := setupTelegramTestDB(t)
	defer db.Close()

	inviteRepo := NewTelegramInviteRepository(db)
	prefRepo := NewPreferensiRepository(db)
	prefID := insertPreference(t, db, 1, "telegram", "777", "granted")

	token, _ := generateToken()
	hash := hashToken(token)
	expires := time.Now().UTC().Add(1 * time.Hour).Format("2006-01-02 15:04:05")
	db.Exec(`INSERT INTO telegram_invite (sekolah_id, preference_id, token_hash, invite_expires_at, telegram_chat_id, status) VALUES (?, ?, ?, ?, 777, 'granted')`,
		1, prefID, hash, expires)

	cfg := TelegramConfig{BotToken: "t", BotUsername: "Bot"}
	svc := NewTelegramService(inviteRepo, prefRepo, cfg)

	var sentText string
	svc.sendFunc = func(_ string, _ int64, text string) error {
		sentText = text
		return nil
	}

	update := map[string]interface{}{
		"message": map[string]interface{}{
			"chat": map[string]interface{}{"id": float64(777), "type": "private"},
			"from": map[string]interface{}{"id": float64(111), "username": "u"},
			"text":  "/status",
		},
	}
	body, _ := json.Marshal(update)

	svc.HandleWebhook(body, "")

	if sentText == "" || sentText != "Status: Terhubung. Anda akan menerima notifikasi." {
		t.Fatalf("expected linked status message, got %q", sentText)
	}
}

func TestWebhookHelp(t *testing.T) {
	db := setupTelegramTestDB(t)
	defer db.Close()

	inviteRepo := NewTelegramInviteRepository(db)
	prefRepo := NewPreferensiRepository(db)

	cfg := TelegramConfig{BotToken: "t", BotUsername: "Bot"}
	svc := NewTelegramService(inviteRepo, prefRepo, cfg)

	var sentText string
	svc.sendFunc = func(_ string, _ int64, text string) error {
		sentText = text
		return nil
	}

	update := map[string]interface{}{
		"message": map[string]interface{}{
			"chat": map[string]interface{}{"id": float64(111), "type": "private"},
			"from": map[string]interface{}{"id": float64(111), "username": "u"},
			"text":  "/help",
		},
	}
	body, _ := json.Marshal(update)

	svc.HandleWebhook(body, "")

	if sentText == "" {
		t.Fatal("expected help text")
	}
}

func TestWebhookSecretValidation(t *testing.T) {
	db := setupTelegramTestDB(t)
	defer db.Close()

	inviteRepo := NewTelegramInviteRepository(db)
	prefRepo := NewPreferensiRepository(db)

	cfg := TelegramConfig{BotToken: "t", BotUsername: "Bot", WebhookSecret: "my-secret"}
	svc := NewTelegramService(inviteRepo, prefRepo, cfg)

	var sentText string
	svc.sendFunc = func(_ string, _ int64, text string) error {
		sentText = text
		return nil
	}

	update := map[string]interface{}{
		"message": map[string]interface{}{
			"chat": map[string]interface{}{"id": float64(111), "type": "private"},
			"from": map[string]interface{}{"id": float64(111), "username": "u"},
			"text":  "/help",
		},
	}
	body, _ := json.Marshal(update)

	err := svc.HandleWebhook(body, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong webhook secret")
	}

	err = svc.HandleWebhook(body, "my-secret")
	if err != nil {
		t.Fatalf("expected success with correct secret: %v", err)
	}

	if sentText == "" {
		t.Fatal("expected help text after correct secret")
	}
}

func TestTenantIsolation(t *testing.T) {
	db := setupTelegramTestDB(t)
	defer db.Close()

	inviteRepo := NewTelegramInviteRepository(db)
	prefRepo := NewPreferensiRepository(db)

	prefID1 := insertPreference(t, db, 1, "telegram", "pending", "pending")
	prefID2 := insertPreference(t, db, 2, "telegram", "pending", "pending")

	cfg := TelegramConfig{BotToken: "t", BotUsername: "Bot"}
	svc := NewTelegramService(inviteRepo, prefRepo, cfg)

	link1, _ := svc.GenerateInviteLink(1, prefID1)
	link2, _ := svc.GenerateInviteLink(2, prefID2)

	token1 := link1[len("https://t.me/Bot?start="):]
	token2 := link2[len("https://t.me/Bot?start="):]

	if token1 == token2 {
		t.Fatal("different tenants should get different tokens")
	}

	hash1 := hashToken(token1)
	inv1, _ := inviteRepo.GetByTokenHash(hash1)
	if inv1.SekolahID != 1 {
		t.Fatalf("expected sekolah_id 1, got %d", inv1.SekolahID)
	}
	if inv1.PreferenceID != prefID1 {
		t.Fatalf("expected preference_id %d, got %d", prefID1, inv1.PreferenceID)
	}
}

func TestProviderSendWithGrantedConsent(t *testing.T) {
	cfg := TelegramConfig{BotToken: "test-token"}
	p := &TelegramProvider{botToken: cfg.BotToken}

	var sentChatID int64
	var sentText string
	p.sendFunc = func(_ string, chatID int64, text string) error {
		sentChatID = chatID
		sentText = text
		return nil
	}

	n := &Notifikasi{Penerima: "12345", Pesan: "Hello"}
	result := p.Send(n)

	if !result.Success {
		t.Fatalf("expected success, got: %v", result.Error)
	}
	if sentChatID != 12345 {
		t.Fatalf("expected chat_id 12345, got %d", sentChatID)
	}
	if sentText != "Hello" {
		t.Fatalf("expected 'Hello', got %q", sentText)
	}
}

func TestProviderSendInvalidChatID(t *testing.T) {
	p := &TelegramProvider{botToken: "t", sendFunc: telegramSendMessage}
	n := &Notifikasi{Penerima: "not-a-number", Pesan: "Hello"}
	result := p.Send(n)

	if result.Success {
		t.Fatal("expected failure for invalid chat_id")
	}
}

func TestProviderSendNoToken(t *testing.T) {
	p := &TelegramProvider{botToken: ""}
	n := &Notifikasi{Penerima: "123", Pesan: "Hello"}
	result := p.Send(n)

	if result.Success {
		t.Fatal("expected failure for missing token")
	}
}

func TestProviderSendBlockedDetection(t *testing.T) {
	p := &TelegramProvider{botToken: "t"}
	p.sendFunc = func(_ string, _ int64, _ string) error {
		return fmt.Errorf("Forbidden: bot was blocked by the user")
	}

	n := &Notifikasi{Penerima: "123", Pesan: "Hello"}
	result := p.Send(n)

	if result.Success {
		t.Fatal("expected failure for blocked user")
	}
	if result.Error == nil || !contains(result.Error.Error(), "blocked/forbidden") {
		t.Fatalf("expected blocked/forbidden error, got: %v", result.Error)
	}
}

func TestProviderSendAPIError(t *testing.T) {
	p := &TelegramProvider{botToken: "t"}
	p.sendFunc = func(_ string, _ int64, _ string) error {
		return fmt.Errorf("Bad Request: chat not found")
	}

	n := &Notifikasi{Penerima: "123", Pesan: "Hello"}
	result := p.Send(n)

	if result.Success {
		t.Fatal("expected failure for api error")
	}
}

func TestIsTelegramBlocked(t *testing.T) {
	tests := []struct {
		msg      string
		expected bool
	}{
		{"Forbidden: bot was blocked by the user", true},
		{"Bad Request: chat not found", true},
		{"Forbidden: user is deactivated", true},
		{"Bad Request: need administrator rights", false},
		{"Request timeout", false},
		{"", false},
	}

	for _, tc := range tests {
		got := isTelegramBlocked(tc.msg)
		if got != tc.expected {
			t.Errorf("isTelegramBlocked(%q) = %v, want %v", tc.msg, got, tc.expected)
		}
	}
}
