package notifikasi

import (
	"fmt"
	"log/slog"
)

type WhatsAppProvider struct{}

func NewWhatsAppProvider() *WhatsAppProvider {
	return &WhatsAppProvider{}
}

func (p *WhatsAppProvider) Type() string {
	return "whatsapp"
}

func (p *WhatsAppProvider) Send(n *Notifikasi) SendResult {
	slog.Info("whatsapp send", "penerima", n.Penerima, "pesan", truncate(n.Pesan, 50))
	return SendResult{
		Success: true,
		Error:   nil,
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + fmt.Sprintf("...(%d chars)", len(s))
}
