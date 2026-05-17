package notifikasi

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Sekolahkit/sekolah-app/pkg/response"
	"github.com/go-chi/chi/v5"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// Phone normalization

var phoneRegex = regexp.MustCompile(`^(\+?62|0)8[1-9][0-9]{6,10}$`)

func NormalizePhone(phone string) (string, error) {
	p := strings.TrimSpace(phone)
	p = strings.ReplaceAll(p, " ", "")
	p = strings.ReplaceAll(p, "-", "")

	if strings.HasPrefix(p, "+62") {
		p = "0" + p[3:]
	}
	if strings.HasPrefix(p, "62") {
		p = "0" + p[2:]
	}

	if !phoneRegex.MatchString(p) {
		return "", fmt.Errorf("nomor telepon tidak valid: %s", phone)
	}

	return p[1:] + "@s.whatsapp.net", nil
}

func PhoneToJID(phone string) (types.JID, error) {
	jidStr, err := NormalizePhone(phone)
	if err != nil {
		return types.JID{}, err
	}
	return types.ParseJID(jidStr)
}

// Rate limiter

type RateLimiterConfig struct {
	SendInterval       time.Duration
	RecipientCooldown  time.Duration
	BurstLimit         int
	BurstWindow        time.Duration
}

func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		SendInterval:      4 * time.Second,
		RecipientCooldown: 30 * time.Second,
		BurstLimit:        20,
		BurstWindow:       5 * time.Minute,
	}
}

type RateLimiter struct {
	cfg          RateLimiterConfig
	mu           sync.Mutex
	lastSend     time.Time
	recipientMap map[string]time.Time
	burstWindow  []time.Time
}

func NewRateLimiter(cfg RateLimiterConfig) *RateLimiter {
	return &RateLimiter{
		cfg:          cfg,
		recipientMap: make(map[string]time.Time),
	}
}

type RateLimitResult struct {
	Allowed  bool
	WaitTime time.Duration
	Reason   string
}

func (rl *RateLimiter) Allow(recipient string) RateLimitResult {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Clean old burst entries
	cutoff := now.Add(-rl.cfg.BurstWindow)
	var valid []time.Time
	for _, t := range rl.burstWindow {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	rl.burstWindow = valid

	// Check burst limit
	if len(rl.burstWindow) >= rl.cfg.BurstLimit {
		wait := rl.burstWindow[0].Add(rl.cfg.BurstWindow).Sub(now)
		return RateLimitResult{Allowed: false, WaitTime: wait, Reason: "burst limit"}
	}

	// Check send interval
	if !rl.lastSend.IsZero() {
		elapsed := now.Sub(rl.lastSend)
		if elapsed < rl.cfg.SendInterval {
			wait := rl.cfg.SendInterval - elapsed
			return RateLimitResult{Allowed: false, WaitTime: wait, Reason: "send interval"}
		}
	}

	// Check per-recipient cooldown
	if lastSent, ok := rl.recipientMap[recipient]; ok {
		elapsed := now.Sub(lastSent)
		if elapsed < rl.cfg.RecipientCooldown {
			wait := rl.cfg.RecipientCooldown - elapsed
			return RateLimitResult{Allowed: false, WaitTime: wait, Reason: "recipient cooldown"}
		}
	}

	// Allow
	rl.lastSend = now
	rl.recipientMap[recipient] = now
	rl.burstWindow = append(rl.burstWindow, now)

	return RateLimitResult{Allowed: true}
}

func (rl *RateLimiter) AddJitter(base time.Duration) time.Duration {
	jitter := time.Duration(rand.Int63n(int64(base / 4)))
	return base + jitter
}

// WhatsApp client wrapper

type WhatsAppClient struct {
	client   *whatsmeow.Client
	qrChan   chan string
	status   string
	lastErr  string
	mu       sync.RWMutex
	logger   *slog.Logger
}

type WhatsAppConfig struct {
	Enabled  bool
	DataPath string
}

func NewWhatsAppClient(cfg WhatsAppConfig) (*WhatsAppClient, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	dbPath := cfg.DataPath
	if dbPath == "" {
		dbPath = "./data/whatsapp.db"
	}

	dbLog := waLog.Stdout("Database", "WARN", true)
	container, err := sqlstore.New(context.Background(), "sqlite3", "file:"+dbPath+"?_journal_mode=WAL", dbLog)
	if err != nil {
		return nil, fmt.Errorf("failed to open whatsapp db: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	clientLog := waLog.Stdout("Client", "WARN", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	wc := &WhatsAppClient{
		client: client,
		qrChan: make(chan string, 1),
		status: "disconnected",
		logger: slog.Default().With("component", "whatsapp-client"),
	}

	client.AddEventHandler(wc.handleEvent)

	return wc, nil
}

func (wc *WhatsAppClient) handleEvent(evt interface{}) {
	switch e := evt.(type) {
	case *events.QR:
		wc.mu.Lock()
		wc.status = "waiting_qr"
		wc.mu.Unlock()
		if len(e.Codes) > 0 {
			select {
			case wc.qrChan <- e.Codes[0]:
			default:
			}
		}
	case *events.Connected:
		wc.mu.Lock()
		wc.status = "connected"
		wc.lastErr = ""
		wc.mu.Unlock()
		wc.logger.Info("WhatsApp connected")
	case *events.Disconnected:
		wc.mu.Lock()
		wc.status = "disconnected"
		wc.mu.Unlock()
		wc.logger.Warn("WhatsApp disconnected")
	}
}

func (wc *WhatsAppClient) Connect(ctx context.Context) error {
	if wc.client.IsConnected() {
		return nil
	}

	if wc.client.Store.ID == nil {
		wc.mu.Lock()
		wc.status = "waiting_qr"
		wc.mu.Unlock()
		qrChan, err := wc.client.GetQRChannel(ctx)
		if err != nil {
			return fmt.Errorf("get QR channel: %w", err)
		}

		go func() {
			for qr := range qrChan {
				select {
				case wc.qrChan <- qr.Code:
				default:
				}
			}
		}()
	}

	err := wc.client.Connect()
	if err != nil {
		wc.mu.Lock()
		wc.lastErr = err.Error()
		wc.mu.Unlock()
		return err
	}

	return nil
}

func (wc *WhatsAppClient) Disconnect() {
	if wc.client != nil {
		wc.client.Disconnect()
		wc.mu.Lock()
		wc.status = "disconnected"
		wc.mu.Unlock()
	}
}

func (wc *WhatsAppClient) IsConnected() bool {
	return wc.client != nil && wc.client.IsConnected()
}

func (wc *WhatsAppClient) StatusInfo() map[string]interface{} {
	wc.mu.RLock()
	defer wc.mu.RUnlock()
	info := map[string]interface{}{
		"status":    wc.status,
		"connected": wc.client != nil && wc.client.IsConnected(),
		"last_error": wc.lastErr,
	}
	if wc.client != nil && wc.client.Store.ID != nil {
		info["logged_in"] = true
	} else {
		info["logged_in"] = false
	}
	return info
}

func (wc *WhatsAppClient) QRChannel() <-chan string {
	return wc.qrChan
}

func (wc *WhatsAppClient) SendText(jid types.JID, text string) error {
	if !wc.IsConnected() {
		return fmt.Errorf("whatsapp not connected")
	}

	msg := &waProto.Message{
		Conversation: &text,
	}

	_, err := wc.client.SendMessage(context.Background(), jid, msg)
	return err
}

// WhatsApp provider

type WhatsAppProvider struct {
	client  *WhatsAppClient
	limiter *RateLimiter
	logger  *slog.Logger
}

func NewWhatsAppProviderWithClient(client *WhatsAppClient, limiter *RateLimiter) *WhatsAppProvider {
	return &WhatsAppProvider{
		client:  client,
		limiter: limiter,
		logger:  slog.Default().With("component", "whatsapp-provider"),
	}
}

func (p *WhatsAppProvider) Type() string {
	return "whatsapp"
}

func (p *WhatsAppProvider) Send(n *Notifikasi) SendResult {
	if p.client == nil || !p.client.IsConnected() {
		return SendResult{
			Success:   false,
			Retryable: true,
			Error:     fmt.Errorf("whatsapp client not connected"),
		}
	}

	jid, err := PhoneToJID(n.Penerima)
	if err != nil {
		return SendResult{
			Success: false,
			Error:   fmt.Errorf("invalid phone: %w", err),
		}
	}

	rlResult := p.limiter.Allow(n.Penerima)
	if !rlResult.Allowed {
		p.logger.Warn("rate limited", "recipient", n.Penerima, "reason", rlResult.Reason, "wait", rlResult.WaitTime)
		return SendResult{
			Success:   false,
			Retryable: true,
			Error:     fmt.Errorf("rate limited (%s), retry in %s", rlResult.Reason, rlResult.WaitTime),
		}
	}

	if err := p.client.SendText(jid, n.Pesan); err != nil {
		errMsg := err.Error()
		if isWhatsAppRetryable(errMsg) {
			return SendResult{
				Success:   false,
				Retryable: true,
				Error:     err,
			}
		}
		return SendResult{
			Success: false,
			Error:   err,
		}
	}

	p.logger.Info("whatsapp sent", "to", n.Penerima)
	return SendResult{Success: true}
}

func isWhatsAppRetryable(errMsg string) bool {
	retryable := []string{
		"not connected",
		"connection closed",
		"timeout",
		"rate limit",
		"too many requests",
		"service unavailable",
	}
	lower := strings.ToLower(errMsg)
	for _, r := range retryable {
		if strings.Contains(lower, r) {
			return true
		}
	}
	return false
}

// WhatsApp handler

type WhatsAppHandler struct {
	client *WhatsAppClient
}

func NewWhatsAppHandler(client *WhatsAppClient) *WhatsAppHandler {
	return &WhatsAppHandler{client: client}
}

func (h *WhatsAppHandler) Status(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		response.JSON(w, 200, map[string]interface{}{
			"status":    "disabled",
			"connected": false,
		})
		return
	}
	response.JSON(w, 200, h.client.StatusInfo())
}

func (h *WhatsAppHandler) Connect(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		response.Error(w, 400, "DISABLED", "WhatsApp tidak dikonfigurasi")
		return
	}

	if h.client.IsConnected() {
		response.JSON(w, 200, map[string]string{"message": "Sudah terhubung"})
		return
	}

	go func() {
		if err := h.client.Connect(context.Background()); err != nil {
			slog.Error("whatsapp connect error", "error", err)
		}
	}()

	response.JSON(w, 200, map[string]string{"message": "Menghubungkan..."})
}

func (h *WhatsAppHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		response.Error(w, 400, "DISABLED", "WhatsApp tidak dikonfigurasi")
		return
	}

	h.client.Disconnect()
	response.JSON(w, 200, map[string]string{"message": "Terputus"})
}

func (h *WhatsAppHandler) QR(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		response.Error(w, 400, "DISABLED", "WhatsApp tidak dikonfigurasi")
		return
	}

	if h.client.IsConnected() {
		response.JSON(w, 200, map[string]interface{}{
			"status": "connected",
			"qr":     nil,
		})
		return
	}

	select {
	case code := <-h.client.QRChannel():
		response.JSON(w, 200, map[string]interface{}{
			"status": "waiting_qr",
			"qr":     code,
		})
	default:
		info := h.client.StatusInfo()
		response.JSON(w, 200, map[string]interface{}{
			"status": info["status"],
			"qr":     nil,
		})
	}
}

func RegisterWhatsAppRoutes(r chi.Router, h *WhatsAppHandler) {
	r.Route("/whatsapp", func(r chi.Router) {
		r.Get("/status", h.Status)
		r.Post("/connect", h.Connect)
		r.Post("/disconnect", h.Disconnect)
		r.Get("/qr", h.QR)
	})
}
