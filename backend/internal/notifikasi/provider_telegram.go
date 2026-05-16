package notifikasi

import (
	"fmt"
	"log/slog"
)

type TelegramProvider struct {
	botToken string
}

func NewTelegramProvider(botToken string) *TelegramProvider {
	return &TelegramProvider{botToken: botToken}
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

	slog.Info("telegram send", "chat_id", n.Penerima, "message", truncate(n.Pesan, 50))
	return SendResult{
		Success: true,
		Error:   nil,
	}
}
