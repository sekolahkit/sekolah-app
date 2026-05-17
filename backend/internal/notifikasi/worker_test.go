package notifikasi

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
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
			claimed_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func setupTestDBMultiConn(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared&_journal_mode=WAL&_busy_timeout=10000")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS notifikasi_antrian (
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
	t.Cleanup(func() {
		db.Exec("DROP TABLE IF EXISTS notifikasi_antrian")
		db.Close()
	})
	return db
}

type mockNotifier struct {
	tipe    string
	results []SendResult
	calls   int
	mu      sync.Mutex
}

func (m *mockNotifier) Send(n *Notifikasi) SendResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.calls < len(m.results) {
		r := m.results[m.calls]
		m.calls++
		return r
	}
	m.calls++
	return SendResult{Success: true}
}

func (m *mockNotifier) Calls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
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

func countByStatus(t *testing.T, db *sql.DB, status string) int {
	t.Helper()
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM notifikasi_antrian WHERE status = ?`, status).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	return count
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

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0, StaleAfter: 5 * time.Minute}
	worker := NewWorker(repo, reg, cfg)
	worker.processBatch(t.Context())

	status, _, _ := getStatus(t, db, id)
	if status != "sent" {
		t.Fatalf("expected status sent, got %s", status)
	}
	if mock.Calls() != 1 {
		t.Fatalf("expected 1 call, got %d", mock.Calls())
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

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0, StaleAfter: 5 * time.Minute}
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

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0, StaleAfter: 5 * time.Minute}
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

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0, StaleAfter: 5 * time.Minute}
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

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0, StaleAfter: 5 * time.Minute}
	worker := NewWorker(repo, reg, cfg)
	worker.processBatch(t.Context())

	if mock.Calls() != 0 {
		t.Fatalf("expected 0 calls (item should be skipped), got %d", mock.Calls())
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

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0, StaleAfter: 5 * time.Minute}
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

	future := time.Now().UTC().Add(1 * time.Hour).Format("2006-01-02 15:04:05")
	_, err := db.Exec(`UPDATE notifikasi_antrian SET scheduled_at = ? WHERE id = ?`, future, id)
	if err != nil {
		t.Fatal(err)
	}

	mock := &mockNotifier{tipe: "email"}
	reg := NewRegistry()
	reg.Register(mock)

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0, StaleAfter: 5 * time.Minute}
	worker := NewWorker(repo, reg, cfg)
	worker.processBatch(t.Context())

	if mock.Calls() != 0 {
		t.Fatalf("expected 0 calls (scheduled for future), got %d", mock.Calls())
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
	if cfg.StaleAfter != 5*time.Minute {
		t.Fatalf("expected stale_after 5m, got %v", cfg.StaleAfter)
	}
}

func TestClaimPendingAtomicity(t *testing.T) {
	db := setupTestDBMultiConn(t)
	defer db.Close()

	repo := NewRepository(db)

	for i := 0; i < 5; i++ {
		insertPending(t, db, 1, "email", fmt.Sprintf("user%d@test.com", i), 0, 3)
	}

	if countByStatus(t, db, "pending") != 5 {
		t.Fatal("expected 5 pending items")
	}

	claimed1, err := repo.ClaimPending(3)
	if err != nil {
		t.Fatalf("first claim failed: %v", err)
	}
	if len(claimed1) != 3 {
		t.Fatalf("expected 3 claimed, got %d", len(claimed1))
	}

	if countByStatus(t, db, "pending") != 2 {
		t.Fatalf("expected 2 remaining pending, got %d", countByStatus(t, db, "pending"))
	}
	if countByStatus(t, db, "processing") != 3 {
		t.Fatalf("expected 3 processing, got %d", countByStatus(t, db, "processing"))
	}

	claimed2, err := repo.ClaimPending(3)
	if err != nil {
		t.Fatalf("second claim failed: %v", err)
	}
	if len(claimed2) != 2 {
		t.Fatalf("expected 2 claimed on second call, got %d", len(claimed2))
	}

	claimed3, err := repo.ClaimPending(3)
	if err != nil {
		t.Fatalf("third claim failed: %v", err)
	}
	if len(claimed3) != 0 {
		t.Fatalf("expected 0 claimed on third call, got %d", len(claimed3))
	}
}

func TestTwoWorkersClaimSameQueue(t *testing.T) {
	db := setupTestDBMultiConn(t)
	defer db.Close()

	totalItems := 4
	for i := 0; i < totalItems; i++ {
		insertPending(t, db, 1, "email", fmt.Sprintf("user%d@test.com", i), 0, 3)
	}

	mock := &mockNotifier{
		tipe: "email",
	}
	reg := NewRegistry()
	reg.Register(mock)

	repo1 := NewRepository(db)
	repo2 := NewRepository(db)

	worker1 := NewWorker(repo1, reg, WorkerConfig{BatchSize: 10, Throttle: 0, StaleAfter: 5 * time.Minute})
	worker2 := NewWorker(repo2, reg, WorkerConfig{BatchSize: 10, Throttle: 0, StaleAfter: 5 * time.Minute})

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		worker1.processBatch(t.Context())
	}()
	go func() {
		defer wg.Done()
		worker2.processBatch(t.Context())
	}()
	wg.Wait()

	sentCount := countByStatus(t, db, "sent")
	pendingCount := countByStatus(t, db, "pending")
	processingCount := countByStatus(t, db, "processing")

	if sentCount != totalItems {
		t.Fatalf("expected %d sent, got %d (pending=%d, processing=%d)",
			totalItems, sentCount, pendingCount, processingCount)
	}
	if pendingCount != 0 {
		t.Fatalf("expected 0 pending, got %d", pendingCount)
	}
	if processingCount != 0 {
		t.Fatalf("expected 0 processing, got %d", processingCount)
	}

	if mock.Calls() != totalItems {
		t.Fatalf("expected %d provider calls, got %d", totalItems, mock.Calls())
	}
}

func TestTwoWorkersNoDuplicateDispatch(t *testing.T) {
	db := setupTestDBMultiConn(t)
	defer db.Close()

	id := insertPending(t, db, 1, "email", "only@test.com", 0, 3)

	callCount := 0
	var mu sync.Mutex
	uniqueNotifier := &trackingNotifier{
		tipe: "email",
		onSend: func(n *Notifikasi) {
			mu.Lock()
			callCount++
			mu.Unlock()
		},
	}
	reg := NewRegistry()
	reg.Register(uniqueNotifier)

	repo1 := NewRepository(db)
	repo2 := NewRepository(db)

	worker1 := NewWorker(repo1, reg, WorkerConfig{BatchSize: 10, Throttle: 0, StaleAfter: 5 * time.Minute})
	worker2 := NewWorker(repo2, reg, WorkerConfig{BatchSize: 10, Throttle: 0, StaleAfter: 5 * time.Minute})

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		worker1.processBatch(t.Context())
	}()
	go func() {
		defer wg.Done()
		worker2.processBatch(t.Context())
	}()
	wg.Wait()

	mu.Lock()
	finalCount := callCount
	mu.Unlock()

	if finalCount != 1 {
		t.Fatalf("expected exactly 1 provider call (no duplicate), got %d", finalCount)
	}

	status, _, _ := getStatus(t, db, id)
	if status != "sent" {
		t.Fatalf("expected status sent, got %s", status)
	}
}

type trackingNotifier struct {
	tipe   string
	onSend func(n *Notifikasi)
}

func (t *trackingNotifier) Send(n *Notifikasi) SendResult {
	t.onSend(n)
	return SendResult{Success: true}
}

func (t *trackingNotifier) Type() string {
	return t.tipe
}

func TestReleaseStale(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	id := insertPending(t, db, 1, "email", "test@example.com", 0, 3)

	_, err := db.Exec(`UPDATE notifikasi_antrian SET status = 'processing', claimed_at = datetime('now', '-10 minutes') WHERE id = ?`, id)
	if err != nil {
		t.Fatal(err)
	}

	status, _, _ := getStatus(t, db, id)
	if status != "processing" {
		t.Fatalf("expected processing, got %s", status)
	}

	released, err := repo.ReleaseStale(5 * time.Minute)
	if err != nil {
		t.Fatalf("ReleaseStale failed: %v", err)
	}
	if released != 1 {
		t.Fatalf("expected 1 released, got %d", released)
	}

	status, _, _ = getStatus(t, db, id)
	if status != "pending" {
		t.Fatalf("expected pending after release, got %s", status)
	}
}

func TestReleaseStaleIgnoresFresh(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	id := insertPending(t, db, 1, "email", "test@example.com", 0, 3)

	_, err := db.Exec(`UPDATE notifikasi_antrian SET status = 'processing', claimed_at = datetime('now') WHERE id = ?`, id)
	if err != nil {
		t.Fatal(err)
	}

	released, err := repo.ReleaseStale(5 * time.Minute)
	if err != nil {
		t.Fatalf("ReleaseStale failed: %v", err)
	}
	if released != 0 {
		t.Fatalf("expected 0 released (too fresh), got %d", released)
	}

	status, _, _ := getStatus(t, db, id)
	if status != "processing" {
		t.Fatalf("expected still processing, got %s", status)
	}
}

func TestWorkerContextCancelReleasesProcessing(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	id1 := insertPending(t, db, 1, "email", "a@test.com", 0, 3)
	insertPending(t, db, 1, "email", "b@test.com", 0, 3)

	callCount := 0
	var mu sync.Mutex
	cancelNotifier := &cancelNotifier{
		tipe: "email",
		onSend: func() {
			mu.Lock()
			callCount++
			if callCount == 1 {
				mu.Unlock()
				return
			}
			mu.Unlock()
		},
	}
	reg := NewRegistry()
	reg.Register(cancelNotifier)

	ctx, cancel := context.WithCancel(t.Context())

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0, StaleAfter: 5 * time.Minute}
	worker := NewWorker(repo, reg, cfg)

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	worker.processBatch(ctx)

	status1, _, _ := getStatus(t, db, id1)
	if status1 != "sent" && status1 != "processing" {
		t.Fatalf("expected first item sent or processing, got %s", status1)
	}
}

type cancelNotifier struct {
	tipe   string
	onSend func()
}

func (c *cancelNotifier) Send(n *Notifikasi) SendResult {
	c.onSend()
	return SendResult{Success: true}
}

func (c *cancelNotifier) Type() string {
	return c.tipe
}
