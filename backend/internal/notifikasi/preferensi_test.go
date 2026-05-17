package notifikasi

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupPreferensiTestDB(t *testing.T) *sql.DB {
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
		CREATE UNIQUE INDEX idx_notifikasi_preferensi_unique ON notifikasi_preferensi(sekolah_id, channel, destination);
		CREATE INDEX idx_notifikasi_preferensi_lookup ON notifikasi_preferensi(sekolah_id, channel, destination, consent_status, enabled);

		CREATE TABLE notifikasi_antrian (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			tipe TEXT NOT NULL,
			penerima TEXT NOT NULL,
			pesan TEXT NOT NULL,
			status TEXT DEFAULT 'pending',
			retry_count INTEGER DEFAULT 0,
			max_retries INTEGER DEFAULT 3,
			last_error TEXT,
			scheduled_at DATETIME,
			sent_at DATETIME,
			claimed_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func insertPreferensi(t *testing.T, db *sql.DB, sekolahID int64, channel, destination, consentStatus string, enabled bool) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO notifikasi_preferensi (sekolah_id, channel, destination, enabled, consent_status, consent_source) VALUES (?, ?, ?, ?, ?, 'admin')`,
		sekolahID, channel, destination, enabled, consentStatus)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCanSendGranted(t *testing.T) {
	db := setupPreferensiTestDB(t)
	defer db.Close()

	repo := NewPreferensiRepository(db)
	insertPreferensi(t, db, 1, "email", "user@example.com", "granted", true)

	allowed, reason := repo.CanSend(1, "email", "user@example.com")
	if !allowed {
		t.Fatalf("expected allowed, got blocked: %s", reason)
	}
	if reason != "" {
		t.Fatalf("expected empty reason, got %q", reason)
	}
}

func TestCanSendDisabled(t *testing.T) {
	db := setupPreferensiTestDB(t)
	defer db.Close()

	repo := NewPreferensiRepository(db)
	insertPreferensi(t, db, 1, "email", "user@example.com", "granted", false)

	allowed, reason := repo.CanSend(1, "email", "user@example.com")
	if allowed {
		t.Fatal("expected blocked for disabled preference")
	}
	if reason != "channel disabled" {
		t.Fatalf("expected 'channel disabled', got %q", reason)
	}
}

func TestCanSendRevoked(t *testing.T) {
	db := setupPreferensiTestDB(t)
	defer db.Close()

	repo := NewPreferensiRepository(db)
	insertPreferensi(t, db, 1, "email", "user@example.com", "revoked", true)

	allowed, reason := repo.CanSend(1, "email", "user@example.com")
	if allowed {
		t.Fatal("expected blocked for revoked consent")
	}
	if reason != "consent revoked" {
		t.Fatalf("expected 'consent revoked', got %q", reason)
	}
}

func TestCanSendPending(t *testing.T) {
	db := setupPreferensiTestDB(t)
	defer db.Close()

	repo := NewPreferensiRepository(db)
	insertPreferensi(t, db, 1, "email", "user@example.com", "pending", true)

	allowed, reason := repo.CanSend(1, "email", "user@example.com")
	if allowed {
		t.Fatal("expected blocked for pending consent")
	}
	if reason != "consent pending" {
		t.Fatalf("expected 'consent pending', got %q", reason)
	}
}

func TestCanSendNoPreference(t *testing.T) {
	db := setupPreferensiTestDB(t)
	defer db.Close()

	repo := NewPreferensiRepository(db)

	allowed, reason := repo.CanSend(1, "email", "unknown@example.com")
	if allowed {
		t.Fatal("expected blocked for missing preference")
	}
	if reason != "no preference found" {
		t.Fatalf("expected 'no preference found', got %q", reason)
	}
}

func TestCanSendTenantIsolation(t *testing.T) {
	db := setupPreferensiTestDB(t)
	defer db.Close()

	repo := NewPreferensiRepository(db)
	insertPreferensi(t, db, 1, "email", "user@example.com", "granted", true)

	allowed, _ := repo.CanSend(2, "email", "user@example.com")
	if allowed {
		t.Fatal("expected blocked for different tenant")
	}
}

func TestCanSendChannelIsolation(t *testing.T) {
	db := setupPreferensiTestDB(t)
	defer db.Close()

	repo := NewPreferensiRepository(db)
	insertPreferensi(t, db, 1, "email", "user@example.com", "granted", true)

	allowed, reason := repo.CanSend(1, "whatsapp", "user@example.com")
	if allowed {
		t.Fatal("expected blocked for different channel")
	}
	if reason != "no preference found" {
		t.Fatalf("expected 'no preference found', got %q", reason)
	}
}

func TestUpsertCreatesNew(t *testing.T) {
	db := setupPreferensiTestDB(t)
	defer db.Close()

	repo := NewPreferensiRepository(db)
	p := &Preferensi{
		SekolahID:     1,
		RecipientType: "manual",
		Channel:       "email",
		Destination:   "new@example.com",
		Enabled:       true,
		ConsentStatus: "granted",
		ConsentSource: "admin",
	}

	id, err := repo.Upsert(p)
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero ID")
	}

	allowed, _ := repo.CanSend(1, "email", "new@example.com")
	if !allowed {
		t.Fatal("expected allowed after upsert")
	}
}

func TestUpsertUpdatesExisting(t *testing.T) {
	db := setupPreferensiTestDB(t)
	defer db.Close()

	repo := NewPreferensiRepository(db)
	insertPreferensi(t, db, 1, "email", "user@example.com", "granted", true)

	allowed, _ := repo.CanSend(1, "email", "user@example.com")
	if !allowed {
		t.Fatal("expected allowed initially")
	}

	p := &Preferensi{
		SekolahID:     1,
		RecipientType: "manual",
		Channel:       "email",
		Destination:   "user@example.com",
		Enabled:       true,
		ConsentStatus: "revoked",
		ConsentSource: "admin",
	}

	_, err := repo.Upsert(p)
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	allowed, reason := repo.CanSend(1, "email", "user@example.com")
	if allowed {
		t.Fatal("expected blocked after revoking")
	}
	if reason != "consent revoked" {
		t.Fatalf("expected 'consent revoked', got %q", reason)
	}
}

func TestWorkerBlocksWhenConsentRevoked(t *testing.T) {
	db := setupPreferensiTestDB(t)
	defer db.Close()

	notifRepo := NewRepository(db)
	prefRepo := NewPreferensiRepository(db)

	insertPreferensi(t, db, 1, "email", "blocked@example.com", "revoked", true)

	id := insertTestNotif(t, db, 1, "email", "blocked@example.com", 0, 3)

	mock := &mockNotifier{tipe: "email"}
	reg := NewRegistry()
	reg.Register(mock)

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0, StaleAfter: 5 * 60}
	worker := NewWorker(notifRepo, reg, cfg).WithPreferensi(prefRepo)
	worker.processBatch(t.Context())

	if mock.Calls() != 0 {
		t.Fatalf("expected 0 provider calls (consent blocked), got %d", mock.Calls())
	}

	status, _, lastError := getNotifStatus(t, db, id)
	if status != "failed" {
		t.Fatalf("expected status failed, got %s", status)
	}
	if lastError != "consent blocked: consent revoked" {
		t.Fatalf("unexpected last_error: %q", lastError)
	}
}

func TestWorkerBlocksWhenDisabled(t *testing.T) {
	db := setupPreferensiTestDB(t)
	defer db.Close()

	notifRepo := NewRepository(db)
	prefRepo := NewPreferensiRepository(db)

	insertPreferensi(t, db, 1, "email", "disabled@example.com", "granted", false)
	insertTestNotif(t, db, 1, "email", "disabled@example.com", 0, 3)

	mock := &mockNotifier{tipe: "email"}
	reg := NewRegistry()
	reg.Register(mock)

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0, StaleAfter: 5 * 60}
	worker := NewWorker(notifRepo, reg, cfg).WithPreferensi(prefRepo)
	worker.processBatch(t.Context())

	if mock.Calls() != 0 {
		t.Fatalf("expected 0 provider calls (disabled), got %d", mock.Calls())
	}
}

func TestWorkerBlocksWhenNoPreference(t *testing.T) {
	db := setupPreferensiTestDB(t)
	defer db.Close()

	notifRepo := NewRepository(db)
	prefRepo := NewPreferensiRepository(db)

	insertTestNotif(t, db, 1, "email", "unknown@example.com", 0, 3)

	mock := &mockNotifier{tipe: "email"}
	reg := NewRegistry()
	reg.Register(mock)

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0, StaleAfter: 5 * 60}
	worker := NewWorker(notifRepo, reg, cfg).WithPreferensi(prefRepo)
	worker.processBatch(t.Context())

	if mock.Calls() != 0 {
		t.Fatalf("expected 0 provider calls (no preference), got %d", mock.Calls())
	}
}

func TestWorkerSendsWhenConsentGranted(t *testing.T) {
	db := setupPreferensiTestDB(t)
	defer db.Close()

	notifRepo := NewRepository(db)
	prefRepo := NewPreferensiRepository(db)

	insertPreferensi(t, db, 1, "email", "allowed@example.com", "granted", true)
	id := insertTestNotif(t, db, 1, "email", "allowed@example.com", 0, 3)

	mock := &mockNotifier{
		tipe:    "email",
		results: []SendResult{{Success: true}},
	}
	reg := NewRegistry()
	reg.Register(mock)

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0, StaleAfter: 5 * 60}
	worker := NewWorker(notifRepo, reg, cfg).WithPreferensi(prefRepo)
	worker.processBatch(t.Context())

	if mock.Calls() != 1 {
		t.Fatalf("expected 1 provider call, got %d", mock.Calls())
	}

	status, _, _ := getNotifStatus(t, db, id)
	if status != "sent" {
		t.Fatalf("expected status sent, got %s", status)
	}
}

func TestWorkerWithoutPrefRepoSkipsCheck(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	notifRepo := NewRepository(db)
	id := insertPending(t, db, 1, "email", "any@example.com", 0, 3)

	mock := &mockNotifier{
		tipe:    "email",
		results: []SendResult{{Success: true}},
	}
	reg := NewRegistry()
	reg.Register(mock)

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0, StaleAfter: 5 * 60}
	worker := NewWorker(notifRepo, reg, cfg)
	worker.processBatch(t.Context())

	if mock.Calls() != 1 {
		t.Fatalf("expected 1 call (no pref repo = skip check), got %d", mock.Calls())
	}

	status, _, _ := getStatus(t, db, id)
	if status != "sent" {
		t.Fatalf("expected sent, got %s", status)
	}
}

func insertTestNotif(t *testing.T, db *sql.DB, sekolahID int64, tipe, penerima string, retryCount, maxRetries int) int64 {
	t.Helper()
	result, err := db.Exec(`INSERT INTO notifikasi_antrian (sekolah_id, tipe, penerima, pesan, status, retry_count, max_retries) VALUES (?, ?, ?, ?, 'pending', ?, ?)`,
		sekolahID, tipe, penerima, "test message", retryCount, maxRetries)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := result.LastInsertId()
	return id
}

func getNotifStatus(t *testing.T, db *sql.DB, id int64) (string, int, string) {
	t.Helper()
	var status string
	var retryCount int
	var lastError string
	err := db.QueryRow(`SELECT status, retry_count, COALESCE(last_error, '') FROM notifikasi_antrian WHERE id = ?`, id).Scan(&status, &retryCount, &lastError)
	if err != nil {
		t.Fatal(err)
	}
	return status, retryCount, lastError
}
