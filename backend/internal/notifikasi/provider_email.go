package notifikasi

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"strings"
)

type EmailConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

func (c EmailConfig) Addr() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func (c EmailConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("SMTP_HOST wajib diisi")
	}
	if c.Port == "" {
		return fmt.Errorf("SMTP_PORT wajib diisi")
	}
	if c.User == "" {
		return fmt.Errorf("SMTP_USER wajib diisi")
	}
	if c.Password == "" {
		return fmt.Errorf("SMTP_PASSWORD wajib diisi")
	}
	return nil
}

func (c EmailConfig) from() string {
	if c.From != "" {
		return c.From
	}
	return c.User
}

type smtpSendFunc func(addr string, a smtp.Auth, from string, to []string, msg []byte) error

type EmailProvider struct {
	cfg      EmailConfig
	sendFunc smtpSendFunc
}

func NewEmailProvider(cfg EmailConfig) *EmailProvider {
	return &EmailProvider{
		cfg:      cfg,
		sendFunc: smtpSendMail,
	}
}

func (p *EmailProvider) Type() string {
	return "email"
}

func (p *EmailProvider) Send(n *Notifikasi) SendResult {
	if err := p.cfg.Validate(); err != nil {
		return SendResult{Success: false, Error: err}
	}

	from := p.cfg.from()
	to := []string{n.Penerima}
	msg := buildEmail(from, n.Penerima, n.Pesan)

	auth := smtp.PlainAuth("", p.cfg.User, p.cfg.Password, p.cfg.Host)

	err := p.sendFunc(p.cfg.Addr(), auth, from, to, []byte(msg))
	if err != nil {
		slog.Error("email send failed", "to", n.Penerima, "error", err)
		return SendResult{Success: false, Error: fmt.Errorf("smtp send: %w", err)}
	}

	slog.Info("email sent", "to", n.Penerima)
	return SendResult{Success: true}
}

func buildEmail(from, to, body string) string {
	subject := "Notifikasi"
	if len(body) > 2 && body[0] == '[' {
		if end := strings.Index(body, "]"); end > 1 {
			subject = body[1:end]
			rest := strings.TrimSpace(body[end+1:])
			if rest != "" {
				body = rest
			}
		}
	}

	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return b.String()
}

func smtpSendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	host, _, _ := net.SplitHostPort(addr)

	if err := smtp.SendMail(addr, a, from, to, msg); err == nil {
		return nil
	}

	return sendMailTLS(addr, host, a, from, to, msg)
}

func sendMailTLS(addr, host string, a smtp.Auth, from string, to []string, msg []byte) error {
	tlsCfg := &tls.Config{ServerName: host}
	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if a != nil {
		if err = client.Auth(a); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	for _, rcpt := range to {
		if err = client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp rcpt: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}

	return client.Quit()
}
