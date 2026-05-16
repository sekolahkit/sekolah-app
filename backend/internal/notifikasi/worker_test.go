package notifikasi

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
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
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

type mockNotifier struct {
	tipe    string
	results []SendResult
	calls   int
}

func (m *mockNotifier) Send(n *Notifikasi) SendResult {
	if m.calls < len(m.results) {
		r := m.results[m.calls]
		m.calls++
		return r
	}
	m.calls++
	return SendResult{Success: true}
}

func (m *mockNotifier) Type() string {
	return m.tipe
}

func insertPending(t *testing.T, db *sql.DB, sekolahID int64, tipe, penerima string, retryCount, maxRetries int) int64 {
	t.Helper()
	result, err := db.Exec(`INSERT INTO notifikasi_antrian (sekolah_id, tipe, penerima, pesan, status, retry_count, max_retries) VALUES (?, ?, ?, ?, 'pending', ?, ?)`,
		sekolahID, tipe, penerima, "test message", retryCount, maxRetries)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := result.LastInsertId()
	return id
}

func getStatus(t *testing.T, db *sql.DB, id int64) (string, int, string) {
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

func TestRegistryGet(t *testing.T) {
	reg := NewRegistry()
	mock := &mockNotifier{tipe: "email"}
	reg.Register(mock)

	got, err := reg.Get("email")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != mock {
		t.Fatal("expected same mock instance")
	}
}

func TestRegistryGetMissing(t *testing.T) {
	reg := NewRegistry()

	_, err := reg.Get("whatsapp")
	if err == nil {
		t.Fatal("expected error for missing provider")
	}
}

func TestRegistryHas(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&mockNotifier{tipe: "email"})

	if !reg.Has("email") {
		t.Fatal("expected Has email = true")
	}
	if reg.Has("whatsapp") {
		t.Fatal("expected Has whatsapp = false")
	}
}

func TestRegistryTypes(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&mockNotifier{tipe: "email"})
	reg.Register(&mockNotifier{tipe: "telegram"})

	types := reg.Types()
	if len(types) != 2 {
		t.Fatalf("expected 2 types, got %d", len(types))
	}
}

func TestWorkerSuccess(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	id := insertPending(t, db, 1, "email", "test@example.com", 0, 3)

	mock := &mockNotifier{
		tipe:    "email",
		results: []SendResult{{Success: true}},
	}
	reg := NewRegistry()
	reg.Register(mock)

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0}
	worker := NewWorker(repo, reg, cfg)
	worker.processBatch(t.Context())

	status, _, _ := getStatus(t, db, id)
	if status != "sent" {
		t.Fatalf("expected status sent, got %s", status)
	}
	if mock.calls != 1 {
		t.Fatalf("expected 1 call, got %d", mock.calls)
	}
}

func TestWorkerFailureIncrementsRetry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	id := insertPending(t, db, 1, "email", "test@example.com", 0, 3)

	mock := &mockNotifier{
		tipe: "email",
		results: []SendResult{
			{Success: false, Error: fmt.Errorf("connection timeout")},
		},
	}
	reg := NewRegistry()
	reg.Register(mock)

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0}
	worker := NewWorker(repo, reg, cfg)
	worker.processBatch(t.Context())

	status, retryCount, lastError := getStatus(t, db, id)
	if status != "pending" {
		t.Fatalf("expected status pending, got %s", status)
	}
	if retryCount != 1 {
		t.Fatalf("expected retry_count 1, got %d", retryCount)
	}
	if lastError != "connection timeout" {
		t.Fatalf("expected last_error 'connection timeout', got %q", lastError)
	}
}

func TestWorkerMaxRetriesMarksFailed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	id := insertPending(t, db, 1, "email", "test@example.com", 2, 3)

	mock := &mockNotifier{
		tipe: "email",
		results: []SendResult{
			{Success: false, Error: fmt.Errorf("permanent error")},
		},
	}
	reg := NewRegistry()
	reg.Register(mock)

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0}
	worker := NewWorker(repo, reg, cfg)
	worker.processBatch(t.Context())

	status, retryCount, lastError := getStatus(t, db, id)
	if status != "failed" {
		t.Fatalf("expected status failed, got %s", status)
	}
	if retryCount != 2 {
		t.Fatalf("expected retry_count 2, got %d", retryCount)
	}
	if lastError != "max retries reached: permanent error" {
		t.Fatalf("unexpected last_error: %q", lastError)
	}
}

func TestWorkerUnknownProviderMarksFailed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	id := insertPending(t, db, 1, "sms", "08123456789", 0, 3)

	reg := NewRegistry()

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0}
	worker := NewWorker(repo, reg, cfg)
	worker.processBatch(t.Context())

	status, _, lastError := getStatus(t, db, id)
	if status != "failed" {
		t.Fatalf("expected status failed, got %s", status)
	}
	if lastError != `provider error: provider "sms" tidak tersedia` {
		t.Fatalf("unexpected last_error: %q", lastError)
	}
}

func TestWorkerSkipsItemsWithMaxRetriesExhausted(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	id := insertPending(t, db, 1, "email", "test@example.com", 3, 3)

	mock := &mockNotifier{tipe: "email"}
	reg := NewRegistry()
	reg.Register(mock)

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0}
	worker := NewWorker(repo, reg, cfg)
	worker.processBatch(t.Context())

	if mock.calls != 0 {
		t.Fatalf("expected 0 calls (item should be skipped), got %d", mock.calls)
	}

	status, _, _ := getStatus(t, db, id)
	if status != "pending" {
		t.Fatalf("expected status unchanged (pending), got %s", status)
	}
}

func TestWorkerMultipleItems(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	id1 := insertPending(t, db, 1, "email", "a@example.com", 0, 3)
	id2 := insertPending(t, db, 1, "email", "b@example.com", 0, 3)
	id3 := insertPending(t, db, 2, "telegram", "12345", 0, 3)

	mock := &mockNotifier{
		tipe:    "email",
		results: []SendResult{{Success: true}, {Success: true}},
	}
	tgMock := &mockNotifier{
		tipe:    "telegram",
		results: []SendResult{{Success: true}},
	}
	reg := NewRegistry()
	reg.Register(mock)
	reg.Register(tgMock)

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0}
	worker := NewWorker(repo, reg, cfg)
	worker.processBatch(t.Context())

	for _, tc := range []struct {
		id     int64
		expect string
	}{
		{id1, "sent"},
		{id2, "sent"},
		{id3, "sent"},
	} {
		status, _, _ := getStatus(t, db, tc.id)
		if status != tc.expect {
			t.Errorf("id %d: expected %s, got %s", tc.id, tc.expect, status)
		}
	}
}

func TestWorkerRespectsScheduledAt(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	id := insertPending(t, db, 1, "email", "test@example.com", 0, 3)

	future := time.Now().Add(1 * time.Hour).Format("2006-01-02 15:04:05")
	_, err := db.Exec(`UPDATE notifikasi_antrian SET scheduled_at = ? WHERE id = ?`, future, id)
	if err != nil {
		t.Fatal(err)
	}

	mock := &mockNotifier{tipe: "email"}
	reg := NewRegistry()
	reg.Register(mock)

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0}
	worker := NewWorker(repo, reg, cfg)
	worker.processBatch(t.Context())

	if mock.calls != 0 {
		t.Fatalf("expected 0 calls (scheduled for future), got %d", mock.calls)
	}
}

func TestDefaultWorkerConfig(t *testing.T) {
	cfg := DefaultWorkerConfig()
	if cfg.Interval != 30*time.Second {
		t.Fatalf("expected interval 30s, got %v", cfg.Interval)
	}
	if cfg.BatchSize != 10 {
		t.Fatalf("expected batch_size 10, got %d", cfg.BatchSize)
	}
	if cfg.Throttle != 500*time.Millisecond {
		t.Fatalf("expected throttle 500ms, got %v", cfg.Throttle)
	}
}
