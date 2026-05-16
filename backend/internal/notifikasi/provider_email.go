package notifikasi

import (
	"fmt"
	"log/slog"
)

type EmailConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

type EmailProvider struct {
	cfg EmailConfig
}

func NewEmailProvider(cfg EmailConfig) *EmailProvider {
	return &EmailProvider{cfg: cfg}
}

func (p *EmailProvider) Type() string {
	return "email"
}

func (p *EmailProvider) Send(n *Notifikasi) SendResult {
	if p.cfg.Host == "" {
		return SendResult{
			Success: false,
			Error:   fmt.Errorf("SMTP_HOST belum dikonfigurasi"),
		}
	}

	slog.Info("email send", "to", n.Penerima, "subject", truncate(n.Pesan, 50))
	return SendResult{
		Success: true,
		Error:   nil,
	}
}
