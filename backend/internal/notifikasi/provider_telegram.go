package notifikasi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type TelegramProvider struct {
	botToken string
	sendFunc func(token string, chatID int64, text string) error
}

func NewTelegramProvider(botToken string) *TelegramProvider {
	return &TelegramProvider{
		botToken: botToken,
		sendFunc: telegramSendMessage,
	}
}

func (p *TelegramProvider) Type() string {
	return "telegram"
}

func (p *TelegramProvider) Send(n *Notifikasi) SendResult {
	if p.botToken == "" {
		return SendResult{
			Success: false,
			Error:   fmt.Errorf("TELEGRAM_BOT_TOKEN belum dikonfigurasi"),
		}
	}

	chatID, err := strconv.ParseInt(n.Penerima, 10, 64)
	if err != nil {
		return SendResult{
			Success: false,
			Error:   fmt.Errorf("invalid telegram chat_id: %q", n.Penerima),
		}
	}

	if err := p.sendFunc(p.botToken, chatID, n.Pesan); err != nil {
		errMsg := err.Error()
		if isTelegramBlocked(errMsg) {
			return SendResult{
				Success: false,
				Error:   fmt.Errorf("telegram blocked/forbidden: %w", err),
			}
		}
		return SendResult{
			Success: false,
			Error:   fmt.Errorf("telegram send: %w", err),
		}
	}

	slog.Info("telegram sent", "chat_id", chatID)
	return SendResult{Success: true}
}

func isTelegramBlocked(errMsg string) bool {
	blocked := []string{"blocked", "forbidden", "chat not found", "user is deactivated"}
	for _, b := range blocked {
		if strings.Contains(strings.ToLower(errMsg), b) {
			return true
		}
	}
	return false
}

type telegramSendAPIResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

func telegramSendHTTPMessage(botToken string, chatID int64, text string) error {
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

	var apiResp telegramSendAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("telegram api decode: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("telegram api error: %s", apiResp.Description)
	}

	return nil
}
