package notifikasi

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/Sekolahkit/sekolah-app/pkg/middleware"
	"github.com/Sekolahkit/sekolah-app/pkg/response"
	"github.com/go-chi/chi/v5"
)

const (
	telegramInviteTTL    = 24 * time.Hour
	telegramStatusPending  = "pending_invite"
	telegramStatusGranted  = "granted"
	telegramStatusRevoked  = "revoked"
	telegramStatusBlocked  = "blocked"
)

type TelegramInvite struct {
	ID              int64  `json:"id"`
	SekolahID       int64  `json:"sekolah_id"`
	PreferenceID    int64  `json:"preference_id"`
	TokenHash       string `json:"-"`
	InviteExpiresAt string `json:"invite_expires_at"`
	InviteUsedAt    string `json:"invite_used_at"`
	TelegramUserID  int64  `json:"telegram_user_id"`
	TelegramChatID  int64  `json:"telegram_chat_id"`
	TelegramUsername string `json:"telegram_username"`
	Status          string `json:"status"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type TelegramInviteRepository struct {
	db *sql.DB
}

func NewTelegramInviteRepository(db *sql.DB) *TelegramInviteRepository {
	return &TelegramInviteRepository{db: db}
}

func (r *TelegramInviteRepository) Create(inv *TelegramInvite) (int64, error) {
	result, err := sq.Insert("telegram_invite").
		Columns("sekolah_id", "preference_id", "token_hash", "invite_expires_at", "status").
		Values(inv.SekolahID, inv.PreferenceID, inv.TokenHash, inv.InviteExpiresAt, telegramStatusPending).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *TelegramInviteRepository) GetByTokenHash(hash string) (*TelegramInvite, error) {
	var inv TelegramInvite
	err := sq.Select("id", "sekolah_id", "preference_id", "token_hash",
		"invite_expires_at", "COALESCE(invite_used_at,'')",
		"COALESCE(telegram_user_id,0)", "COALESCE(telegram_chat_id,0)",
		"COALESCE(telegram_username,'')", "status",
		"created_at", "COALESCE(updated_at,created_at)").
		From("telegram_invite").
		Where(sq.Eq{"token_hash": hash}).
		RunWith(r.db).QueryRow().
		Scan(&inv.ID, &inv.SekolahID, &inv.PreferenceID, &inv.TokenHash,
			&inv.InviteExpiresAt, &inv.InviteUsedAt,
			&inv.TelegramUserID, &inv.TelegramChatID,
			&inv.TelegramUsername, &inv.Status,
			&inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *TelegramInviteRepository) GetByChatID(chatID int64) (*TelegramInvite, error) {
	var inv TelegramInvite
	err := sq.Select("id", "sekolah_id", "preference_id", "token_hash",
		"invite_expires_at", "COALESCE(invite_used_at,'')",
		"COALESCE(telegram_user_id,0)", "COALESCE(telegram_chat_id,0)",
		"COALESCE(telegram_username,'')", "status",
		"created_at", "COALESCE(updated_at,created_at)").
		From("telegram_invite").
		Where(sq.Eq{"telegram_chat_id": chatID}).
		Where(sq.NotEq{"status": telegramStatusPending}).
		OrderBy("updated_at DESC").
		RunWith(r.db).QueryRow().
		Scan(&inv.ID, &inv.SekolahID, &inv.PreferenceID, &inv.TokenHash,
			&inv.InviteExpiresAt, &inv.InviteUsedAt,
			&inv.TelegramUserID, &inv.TelegramChatID,
			&inv.TelegramUsername, &inv.Status,
			&inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *TelegramInviteRepository) MarkUsed(id int64, userID, chatID int64, username string) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	_, err := sq.Update("telegram_invite").
		Set("invite_used_at", now).
		Set("telegram_user_id", userID).
		Set("telegram_chat_id", chatID).
		Set("telegram_username", username).
		Set("status", telegramStatusGranted).
		Set("updated_at", now).
		Where(sq.Eq{"id": id}).
		RunWith(r.db).Exec()
	return err
}

func (r *TelegramInviteRepository) UpdateStatus(id int64, status string) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	_, err := sq.Update("telegram_invite").
		Set("status", status).
		Set("updated_at", now).
		Where(sq.Eq{"id": id}).
		RunWith(r.db).Exec()
	return err
}

type TelegramConfig struct {
	BotToken      string
	BotUsername   string
	WebhookSecret string
}

type TelegramService struct {
	inviteRepo  *TelegramInviteRepository
	prefRepo    *PreferensiRepository
	cfg         TelegramConfig
	sendFunc    func(token string, chatID int64, text string) error
}

func NewTelegramService(
	inviteRepo *TelegramInviteRepository,
	prefRepo *PreferensiRepository,
	cfg TelegramConfig,
) *TelegramService {
	return &TelegramService{
		inviteRepo: inviteRepo,
		prefRepo:   prefRepo,
		cfg:        cfg,
		sendFunc:   telegramSendMessage,
	}
}

func (s *TelegramService) GenerateInviteLink(sekolahID, preferenceID int64) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	hash := hashToken(token)

	expires := time.Now().UTC().Add(telegramInviteTTL).Format("2006-01-02 15:04:05")
	inv := &TelegramInvite{
		SekolahID:       sekolahID,
		PreferenceID:    preferenceID,
		TokenHash:       hash,
		InviteExpiresAt: expires,
	}

	if _, err := s.inviteRepo.Create(inv); err != nil {
		return "", fmt.Errorf("create invite: %w", err)
	}

	link := fmt.Sprintf("https://t.me/%s?start=%s", s.cfg.BotUsername, token)
	return link, nil
}

func (s *TelegramService) HandleWebhook(body []byte, secretHeader string) error {
	if s.cfg.WebhookSecret != "" && secretHeader != s.cfg.WebhookSecret {
		return fmt.Errorf("invalid webhook secret")
	}

	var update struct {
		Message *struct {
			Chat struct {
				ID    int64  `json:"id"`
				Type  string `json:"type"`
			} `json:"chat"`
			From struct {
				ID       int64  `json:"id"`
				Username string `json:"username"`
			} `json:"from"`
			Text string `json:"text"`
		} `json:"message"`
	}

	if err := json.Unmarshal(body, &update); err != nil {
		return fmt.Errorf("parse update: %w", err)
	}

	if update.Message == nil {
		return nil
	}

	msg := update.Message
	chatID := msg.Chat.ID
	userID := msg.From.ID
	username := msg.From.Username
	text := strings.TrimSpace(msg.Text)

	switch {
	case strings.HasPrefix(text, "/start"):
		return s.handleStart(chatID, userID, username, text)
	case text == "/status":
		return s.handleStatus(chatID)
	case text == "/stop":
		return s.handleStop(chatID)
	case text == "/help":
		return s.handleHelp(chatID)
	default:
		return nil
	}
}

func (s *TelegramService) handleStart(chatID, userID int64, username, text string) error {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		return s.sendMsg(chatID, "Selamat datang! Kirim /start <token> untuk menghubungkan akun Anda.")
	}

	token := parts[1]
	hash := hashToken(token)

	inv, err := s.inviteRepo.GetByTokenHash(hash)
	if err != nil {
		slog.Warn("telegram /start: token not found", "chat_id", chatID)
		return s.sendMsg(chatID, "Token tidak valid atau sudah kedaluwarsa.")
	}

	if inv.Status != telegramStatusPending {
		return s.sendMsg(chatID, "Token ini sudah digunakan.")
	}

	if time.Now().UTC().After(parseTime(inv.InviteExpiresAt)) {
		s.inviteRepo.UpdateStatus(inv.ID, telegramStatusRevoked)
		return s.sendMsg(chatID, "Token ini sudah kedaluwarsa. Minta admin untuk membuat undangan baru.")
	}

	if err := s.inviteRepo.MarkUsed(inv.ID, userID, chatID, username); err != nil {
		return fmt.Errorf("mark invite used: %w", err)
	}

	if err := s.updatePreferenceChatID(inv.PreferenceID, chatID); err != nil {
		slog.Error("telegram: failed to update preference", "error", err)
	}

	slog.Info("telegram linked", "chat_id", chatID, "preference_id", inv.PreferenceID)
	return s.sendMsg(chatID, "Berhasil terhubung! Anda akan menerima notifikasi melalui Telegram.")
}

func (s *TelegramService) handleStatus(chatID int64) error {
	inv, err := s.inviteRepo.GetByChatID(chatID)
	if err != nil {
		return s.sendMsg(chatID, "Chat ini belum terhubung. Gunakan /start <token> untuk menghubungkan.")
	}

	switch inv.Status {
	case telegramStatusGranted:
		return s.sendMsg(chatID, "Status: Terhubung. Anda akan menerima notifikasi.")
	case telegramStatusRevoked:
		return s.sendMsg(chatID, "Status: Dicabut. Hubungi admin untuk membuat undangan baru.")
	case telegramStatusBlocked:
		return s.sendMsg(chatID, "Status: Diblokir.")
	default:
		return s.sendMsg(chatID, "Status: Tidak diketahui.")
	}
}

func (s *TelegramService) handleStop(chatID int64) error {
	inv, err := s.inviteRepo.GetByChatID(chatID)
	if err != nil {
		return s.sendMsg(chatID, "Chat ini belum terhubung.")
	}

	if err := s.inviteRepo.UpdateStatus(inv.ID, telegramStatusRevoked); err != nil {
		return fmt.Errorf("revoke invite: %w", err)
	}

	if err := s.updatePreferenceConsent(inv.PreferenceID, "revoked"); err != nil {
		slog.Error("telegram: failed to revoke preference consent", "error", err)
	}

	slog.Info("telegram stopped", "chat_id", chatID, "preference_id", inv.PreferenceID)
	return s.sendMsg(chatID, "Anda telah berhenti menerima notifikasi Telegram. Kirim /start <token> untuk menghubungkan kembali.")
}

func (s *TelegramService) handleHelp(chatID int64) error {
	text := "Perintah yang tersedia:\n\n" +
		"/start <token> - Hubungkan akun penerima notifikasi\n" +
		"/status - Cek status koneksi\n" +
		"/stop - Berhenti menerima notifikasi\n" +
		"/help - Tampilkan bantuan ini"
	return s.sendMsg(chatID, text)
}

func (s *TelegramService) updatePreferenceChatID(prefID, chatID int64) error {
	_, err := sq.Update("notifikasi_preferensi").
		Set("destination", fmt.Sprintf("%d", chatID)).
		Set("consent_status", "granted").
		Set("consent_source", "self_service").
		Set("consent_at", time.Now().UTC().Format("2006-01-02 15:04:05")).
		Set("updated_at", time.Now().UTC().Format("2006-01-02 15:04:05")).
		Where(sq.Eq{"id": prefID}).
		RunWith(s.prefRepo.db).Exec()
	return err
}

func (s *TelegramService) updatePreferenceConsent(prefID int64, status string) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	builder := sq.Update("notifikasi_preferensi").
		Set("consent_status", status).
		Set("consent_source", "self_service").
		Set("updated_at", now).
		Where(sq.Eq{"id": prefID})
	if status == "revoked" {
		builder = builder.Set("revoked_at", now)
	}
	_, err := builder.RunWith(s.prefRepo.db).Exec()
	return err
}

func (s *TelegramService) sendMsg(chatID int64, text string) error {
	if s.sendFunc == nil {
		return nil
	}
	return s.sendFunc(s.cfg.BotToken, chatID, text)
}

type TelegramHandler struct {
	service *TelegramService
}

func NewTelegramHandler(service *TelegramService) *TelegramHandler {
	return &TelegramHandler{service: service}
}

func (h *TelegramHandler) GenerateInvite(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	idStr := chi.URLParam(r, "id")
	prefID, err := parseID(idStr)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID preferensi tidak valid")
		return
	}

	link, err := h.service.GenerateInviteLink(user.SekolahID, prefID)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal membuat link undangan")
		return
	}

	response.JSON(w, 200, map[string]string{
		"invite_link": link,
		"expires_in":  fmt.Sprintf("%d hours", int(telegramInviteTTL.Hours())),
	})
}

func (h *TelegramHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, 400, "BAD_REQUEST", "Gagal membaca body")
		return
	}

	secretHeader := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")

	if err := h.service.HandleWebhook(body, secretHeader); err != nil {
		slog.Error("telegram webhook error", "error", err)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"ok":true}`)
}

func RegisterTelegramRoutes(r chi.Router, h *TelegramHandler) {
	r.Post("/notifikasi/preferensi/{id}/telegram-invite", h.GenerateInvite)
}

func RegisterTelegramWebhook(public chi.Router, h *TelegramHandler) {
	public.Post("/telegram/webhook", h.Webhook)
}

func generateToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func parseTime(s string) time.Time {
	formats := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05Z",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func parseID(s string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(s, "%d", &id)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}

type telegramAPIResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

func telegramSendMessage(botToken string, chatID int64, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("telegram api: %w", err)
	}
	defer resp.Body.Close()

	var apiResp telegramAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("telegram api decode: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("telegram api error: %s", apiResp.Description)
	}

	return nil
}
