package notifikasi

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"081234567890", "81234567890@s.whatsapp.net", false},
		{"+6281234567890", "81234567890@s.whatsapp.net", false},
		{"6281234567890", "81234567890@s.whatsapp.net", false},
		{"0812 3456 7890", "81234567890@s.whatsapp.net", false},
		{"0812-3456-7890", "81234567890@s.whatsapp.net", false},
		{"+62 812 3456 7890", "81234567890@s.whatsapp.net", false},
		{"08123456789", "8123456789@s.whatsapp.net", false},
		{"08123456789012", "", true},
		{"0712345678", "", true},
		{"123", "", true},
		{"", "", true},
		{"abcdefghij", "", true},
		{"+1234567890", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := NormalizePhone(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q, got %q", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("NormalizePhone(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestRateLimiterAllow(t *testing.T) {
	cfg := RateLimiterConfig{
		SendInterval:      50 * time.Millisecond,
		RecipientCooldown: 100 * time.Millisecond,
		BurstLimit:        5,
		BurstWindow:       1 * time.Second,
	}
	rl := NewRateLimiter(cfg)

	r1 := rl.Allow("user1")
	if !r1.Allowed {
		t.Fatalf("first call should be allowed, got: %s (wait: %s)", r1.Reason, r1.WaitTime)
	}

	r2 := rl.Allow("user1")
	if r2.Allowed {
		t.Fatal("second call immediately should be blocked (send interval)")
	}
	if r2.Reason != "send interval" {
		t.Fatalf("expected send interval, got %s", r2.Reason)
	}

	time.Sleep(60 * time.Millisecond)

	r3 := rl.Allow("user2")
	if !r3.Allowed {
		t.Fatalf("different recipient after interval should be allowed, got: %s", r3.Reason)
	}

	r4 := rl.Allow("user1")
	if r4.Allowed {
		t.Fatal("should be blocked by recipient cooldown")
	}
}

func TestRateLimiterRecipientCooldown(t *testing.T) {
	cfg := RateLimiterConfig{
		SendInterval:      10 * time.Millisecond,
		RecipientCooldown: 100 * time.Millisecond,
		BurstLimit:        100,
		BurstWindow:       1 * time.Second,
	}
	rl := NewRateLimiter(cfg)

	rl.Allow("user1")
	time.Sleep(15 * time.Millisecond)

	r := rl.Allow("user1")
	if r.Allowed {
		t.Fatal("should be blocked by recipient cooldown")
	}
	if r.Reason != "recipient cooldown" {
		t.Fatalf("expected recipient cooldown, got %s", r.Reason)
	}

	r2 := rl.Allow("user2")
	if !r2.Allowed {
		t.Fatalf("different recipient should be allowed, got: %s", r2.Reason)
	}
}

func TestRateLimiterBurstLimit(t *testing.T) {
	cfg := RateLimiterConfig{
		SendInterval:      1 * time.Millisecond,
		RecipientCooldown: 1 * time.Millisecond,
		BurstLimit:        3,
		BurstWindow:       1 * time.Second,
	}
	rl := NewRateLimiter(cfg)

	for i := 0; i < 3; i++ {
		time.Sleep(2 * time.Millisecond)
		r := rl.Allow(fmt.Sprintf("user%d", i))
		if !r.Allowed {
			t.Fatalf("call %d should be allowed, got: %s", i, r.Reason)
		}
	}

	time.Sleep(2 * time.Millisecond)
	r := rl.Allow("user99")
	if r.Allowed {
		t.Fatal("should be blocked by burst limit")
	}
	if r.Reason != "burst limit" {
		t.Fatalf("expected burst limit, got %s", r.Reason)
	}
}

func TestRateLimiterAddJitter(t *testing.T) {
	rl := NewRateLimiter(DefaultRateLimiterConfig())
	base := 10 * time.Second
	result := rl.AddJitter(base)

	if result < base {
		t.Fatalf("jitter result %v should be >= base %v", result, base)
	}
	if result > base+base/4 {
		t.Fatalf("jitter result %v should be <= base + 25%% (%v)", result, base+base/4)
	}
}

func TestProviderSendNotConnected(t *testing.T) {
	p := &WhatsAppProvider{client: nil, limiter: NewRateLimiter(DefaultRateLimiterConfig())}
	n := &Notifikasi{Penerima: "081234567890", Pesan: "Test"}
	result := p.Send(n)

	if result.Success {
		t.Fatal("expected failure when not connected")
	}
	if !result.Retryable {
		t.Fatal("expected retryable when not connected")
	}
}

func TestProviderSendInvalidPhone(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{
		SendInterval:      1 * time.Millisecond,
		RecipientCooldown: 1 * time.Millisecond,
		BurstLimit:        100,
		BurstWindow:       1 * time.Minute,
	})

	p := &WhatsAppProvider{client: nil, limiter: rl}
	n := &Notifikasi{Penerima: "invalid", Pesan: "Test"}
	result := p.Send(n)

	if result.Success {
		t.Fatal("expected failure for invalid phone")
	}
	if !result.Retryable {
		t.Fatal("not connected should be retryable")
	}
}

func TestProviderRateLimited(t *testing.T) {
	client := &WhatsAppClient{status: "connected"}
	client.client = nil

	rl := NewRateLimiter(RateLimiterConfig{
		SendInterval:      1 * time.Hour,
		RecipientCooldown: 1 * time.Hour,
		BurstLimit:        1000,
		BurstWindow:       1 * time.Hour,
	})

	p := &WhatsAppProvider{client: client, limiter: rl}
	n := &Notifikasi{Penerima: "081234567890", Pesan: "Test"}

	rl.lastSend = time.Now()
	rl.recipientMap["081234567890"] = time.Now()

	result := p.Send(n)
	if result.Success {
		t.Fatal("expected failure when rate limited")
	}
	if !result.Retryable {
		t.Fatal("rate limited should be retryable")
	}
}

func TestWorkerRescheduleOnRetryable(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	id := insertPending(t, db, 1, "whatsapp", "081234567890", 0, 3)

	rl := NewRateLimiter(RateLimiterConfig{
		SendInterval:      1 * time.Hour,
		RecipientCooldown: 1 * time.Hour,
		BurstLimit:        1000,
		BurstWindow:       1 * time.Hour,
	})
	rl.lastSend = time.Now()
	rl.recipientMap["081234567890"] = time.Now()

	client := &WhatsAppClient{status: "connected"}
	p := &WhatsAppProvider{client: client, limiter: rl}

	reg := NewRegistry()
	reg.Register(p)

	cfg := WorkerConfig{BatchSize: 10, Throttle: 0, StaleAfter: 5 * time.Minute, RetryDelay: 30 * time.Second}
	worker := NewWorker(repo, reg, cfg)
	worker.processBatch(t.Context())

	status, retryCount, lastError := getStatus(t, db, id)
	if status != "pending" {
		t.Fatalf("expected status pending after reschedule, got %s", status)
	}
	if retryCount != 0 {
		t.Fatalf("expected retry_count unchanged (0), got %d", retryCount)
	}
	if lastError == "" {
		t.Fatal("expected last_error to be set")
	}

	var scheduledAt sql.NullString
	db.QueryRow(`SELECT scheduled_at FROM notifikasi_antrian WHERE id = ?`, id).Scan(&scheduledAt)
	if !scheduledAt.Valid || scheduledAt.String == "" {
		t.Fatal("expected scheduled_at to be set after reschedule")
	}
}

func TestRescheduleSetsScheduledAt(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	id := insertPending(t, db, 1, "whatsapp", "081234567890", 0, 3)

	future := time.Now().UTC().Add(5 * time.Minute).Format("2006-01-02 15:04:05")
	err := repo.Reschedule(id, future, "rate limited")
	if err != nil {
		t.Fatalf("Reschedule failed: %v", err)
	}

	status, _, lastError := getStatus(t, db, id)
	if status != "pending" {
		t.Fatalf("expected pending, got %s", status)
	}
	if lastError != "rate limited" {
		t.Fatalf("expected 'rate limited', got %q", lastError)
	}
}

func TestIsWhatsAppRetryable(t *testing.T) {
	tests := []struct {
		msg      string
		expected bool
	}{
		{"not connected", true},
		{"connection closed", true},
		{"timeout", true},
		{"rate limit exceeded", true},
		{"too many requests", true},
		{"service unavailable", true},
		{"invalid JID", false},
		{"recipient not on whatsapp", false},
		{"", false},
	}

	for _, tc := range tests {
		got := isWhatsAppRetryable(tc.msg)
		if got != tc.expected {
			t.Errorf("isWhatsAppRetryable(%q) = %v, want %v", tc.msg, got, tc.expected)
		}
	}
}
