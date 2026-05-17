package notifikasi

import (
	"fmt"
	"net/smtp"
	"sync"
	"testing"
)

func TestEmailConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     EmailConfig
		wantErr string
	}{
		{
			name:    "empty host",
			cfg:     EmailConfig{Port: "587", User: "u", Password: "p"},
			wantErr: "SMTP_HOST wajib diisi",
		},
		{
			name:    "empty port",
			cfg:     EmailConfig{Host: "smtp.example.com", User: "u", Password: "p"},
			wantErr: "SMTP_PORT wajib diisi",
		},
		{
			name:    "empty user",
			cfg:     EmailConfig{Host: "smtp.example.com", Port: "587", Password: "p"},
			wantErr: "SMTP_USER wajib diisi",
		},
		{
			name:    "empty password",
			cfg:     EmailConfig{Host: "smtp.example.com", Port: "587", User: "u"},
			wantErr: "SMTP_PASSWORD wajib diisi",
		},
		{
			name: "valid config",
			cfg:  EmailConfig{Host: "smtp.example.com", Port: "587", User: "u", Password: "p"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tc.wantErr)
				}
				if err.Error() != tc.wantErr {
					t.Fatalf("expected %q, got %q", tc.wantErr, err.Error())
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestEmailConfigFrom(t *testing.T) {
	cfg := EmailConfig{User: "user@example.com"}
	if cfg.from() != "user@example.com" {
		t.Fatalf("expected fallback to User, got %q", cfg.from())
	}

	cfg.From = "noreply@example.com"
	if cfg.from() != "noreply@example.com" {
		t.Fatalf("expected From, got %q", cfg.from())
	}
}

func TestEmailConfigAddr(t *testing.T) {
	cfg := EmailConfig{Host: "smtp.example.com", Port: "587"}
	if cfg.Addr() != "smtp.example.com:587" {
		t.Fatalf("expected smtp.example.com:587, got %q", cfg.Addr())
	}
}

type mockSendFunc struct {
	mu      sync.Mutex
	calls   []mockSendCall
	results []SendResult
}

type mockSendCall struct {
	addr string
	from string
	to   []string
	msg  string
}

func (m *mockSendFunc) send(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, mockSendCall{
		addr: addr,
		from: from,
		to:   to,
		msg:  string(msg),
	})
	idx := len(m.calls) - 1
	if idx < len(m.results) {
		r := m.results[idx]
		if r.Error != nil {
			return r.Error
		}
		return nil
	}
	return nil
}

func (m *mockSendFunc) Calls() []mockSendCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]mockSendCall, len(m.calls))
	copy(cp, m.calls)
	return cp
}

func TestEmailProviderSendSuccess(t *testing.T) {
	mock := &mockSendFunc{
		results: []SendResult{{Success: true}},
	}
	cfg := EmailConfig{
		Host:     "smtp.example.com",
		Port:     "587",
		User:     "sender@example.com",
		Password: "password",
	}
	p := &EmailProvider{cfg: cfg, sendFunc: mock.send}

	n := &Notifikasi{
		Penerima: "recipient@example.com",
		Pesan:    "Hello World",
	}

	result := p.Send(n)
	if !result.Success {
		t.Fatalf("expected success, got error: %v", result.Error)
	}

	calls := mock.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].from != "sender@example.com" {
		t.Fatalf("expected from sender@example.com, got %q", calls[0].from)
	}
	if len(calls[0].to) != 1 || calls[0].to[0] != "recipient@example.com" {
		t.Fatalf("unexpected to: %v", calls[0].to)
	}
	if calls[0].addr != "smtp.example.com:587" {
		t.Fatalf("unexpected addr: %q", calls[0].addr)
	}
}

func TestEmailProviderSendWithFrom(t *testing.T) {
	mock := &mockSendFunc{
		results: []SendResult{{Success: true}},
	}
	cfg := EmailConfig{
		Host:     "smtp.example.com",
		Port:     "587",
		User:     "user@example.com",
		Password: "password",
		From:     "noreply@sekolah.com",
	}
	p := &EmailProvider{cfg: cfg, sendFunc: mock.send}

	n := &Notifikasi{Penerima: "to@example.com", Pesan: "Test"}
	result := p.Send(n)
	if !result.Success {
		t.Fatalf("expected success, got error: %v", result.Error)
	}

	calls := mock.Calls()
	if calls[0].from != "noreply@sekolah.com" {
		t.Fatalf("expected from noreply@sekolah.com, got %q", calls[0].from)
	}
}

func TestEmailProviderSendSMTPFailure(t *testing.T) {
	mock := &mockSendFunc{
		results: []SendResult{
			{Success: false, Error: fmt.Errorf("connection refused")},
		},
	}
	cfg := EmailConfig{
		Host:     "smtp.example.com",
		Port:     "587",
		User:     "sender@example.com",
		Password: "password",
	}
	p := &EmailProvider{cfg: cfg, sendFunc: mock.send}

	n := &Notifikasi{Penerima: "to@example.com", Pesan: "Test"}
	result := p.Send(n)
	if result.Success {
		t.Fatal("expected failure, got success")
	}
	if result.Error == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestEmailProviderSendInvalidConfig(t *testing.T) {
	mock := &mockSendFunc{}
	cfg := EmailConfig{Host: "", Port: "587", User: "u", Password: "p"}
	p := &EmailProvider{cfg: cfg, sendFunc: mock.send}

	n := &Notifikasi{Penerima: "to@example.com", Pesan: "Test"}
	result := p.Send(n)
	if result.Success {
		t.Fatal("expected failure for invalid config")
	}
	if result.Error == nil || result.Error.Error() != "SMTP_HOST wajib diisi" {
		t.Fatalf("expected SMTP_HOST error, got: %v", result.Error)
	}

	calls := mock.Calls()
	if len(calls) != 0 {
		t.Fatalf("expected 0 calls (config validation failed), got %d", len(calls))
	}
}

func TestEmailProviderWorkerRetryFlow(t *testing.T) {
	mock := &mockSendFunc{
		results: []SendResult{
			{Success: false, Error: fmt.Errorf("temporary failure")},
			{Success: false, Error: fmt.Errorf("temporary failure")},
			{Success: true},
		},
	}
	cfg := EmailConfig{
		Host:     "smtp.example.com",
		Port:     "587",
		User:     "sender@example.com",
		Password: "password",
	}
	p := &EmailProvider{cfg: cfg, sendFunc: mock.send}

	n := &Notifikasi{Penerima: "to@example.com", Pesan: "Test", RetryCount: 0, MaxRetries: 3}

	result1 := p.Send(n)
	if result1.Success {
		t.Fatal("expected first call to fail")
	}

	result2 := p.Send(n)
	if result2.Success {
		t.Fatal("expected second call to fail")
	}

	result3 := p.Send(n)
	if !result3.Success {
		t.Fatalf("expected third call to succeed, got: %v", result3.Error)
	}

	if len(mock.Calls()) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(mock.Calls()))
	}
}

func TestBuildEmail(t *testing.T) {
	msg := buildEmail("from@example.com", "to@example.com", "Hello World")
	if msg == "" {
		t.Fatal("expected non-empty message")
	}
	if !contains(msg, "From: from@example.com") {
		t.Fatal("missing From header")
	}
	if !contains(msg, "To: to@example.com") {
		t.Fatal("missing To header")
	}
	if !contains(msg, "Subject: Notifikasi") {
		t.Fatal("missing Subject header")
	}
	if !contains(msg, "Hello World") {
		t.Fatal("missing body")
	}
	if !contains(msg, "Content-Type: text/plain") {
		t.Fatal("missing Content-Type")
	}
}

func TestBuildEmailWithSubject(t *testing.T) {
	msg := buildEmail("from@example.com", "to@example.com", "[Tagihan SPP] Pembayaran bulan Mei")
	if !contains(msg, "Subject: Tagihan SPP") {
		t.Fatalf("expected subject from brackets, got: %s", msg)
	}
	if !contains(msg, "Pembayaran bulan Mei") {
		t.Fatal("missing body after subject extraction")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
