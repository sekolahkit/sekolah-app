package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App       AppConfig       `yaml:"app"`
	Database  DatabaseConfig  `yaml:"database"`
	Modules   ModulesConfig   `yaml:"modules"`
	Notifikasi NotifikasiConfig `yaml:"notifikasi"`
	Upload    UploadConfig    `yaml:"upload"`
	Backup    BackupConfig    `yaml:"backup"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	CORS      CORSConfig      `yaml:"cors"`
	Logging   LoggingConfig   `yaml:"logging"`
}

type AppConfig struct {
	Port    int    `yaml:"port"`
	Name    string `yaml:"name"`
	BaseURL string `yaml:"base_url"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type ModulesConfig struct {
	PPDB       bool `yaml:"ppdb"`
	Pembayaran bool `yaml:"pembayaran"`
	Notifikasi bool `yaml:"notifikasi"`
}

type NotifikasiConfig struct {
	WhatsApp bool `yaml:"whatsapp"`
	Telegram bool `yaml:"telegram"`
	Email    bool `yaml:"email"`
}

type UploadConfig struct {
	MaxSize      int      `yaml:"max_size"`
	AllowedTypes []string `yaml:"allowed_types"`
}

type BackupConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Schedule  string `yaml:"schedule"`
	Retention int    `yaml:"retention"`
	Path      string `yaml:"path"`
}

type RateLimitConfig struct {
	Login    int `yaml:"login"`
	Register int `yaml:"register"`
	API      int `yaml:"api"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
	File   string `yaml:"file"`
}

type Secrets struct {
	JWTSecret          string
	GoogleClientID     string
	GoogleClientSecret string
	SMTPHost           string
	SMTPPort           string
	SMTPUser           string
	SMTPPassword       string
	SMTPFrom           string
	TelegramBotToken    string
	TelegramBotUsername  string
	TelegramWebhookSecret string
	MidtransServerKey  string
	MidtransClientKey  string
	XenditSecretKey    string
}

func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	applyDefaults(cfg)

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}

func LoadSecrets() *Secrets {
	return &Secrets{
		JWTSecret:          os.Getenv("JWT_SECRET"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		SMTPHost:           os.Getenv("SMTP_HOST"),
		SMTPPort:           os.Getenv("SMTP_PORT"),
		SMTPUser:           os.Getenv("SMTP_USER"),
		SMTPPassword:       os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:           os.Getenv("SMTP_FROM"),
		TelegramBotToken:    os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramBotUsername:  os.Getenv("TELEGRAM_BOT_USERNAME"),
		TelegramWebhookSecret: os.Getenv("TELEGRAM_WEBHOOK_SECRET"),
		MidtransServerKey:  os.Getenv("MIDTRANS_SERVER_KEY"),
		MidtransClientKey:  os.Getenv("MIDTRANS_CLIENT_KEY"),
		XenditSecretKey:    os.Getenv("XENDIT_SECRET_KEY"),
	}
}

func applyDefaults(cfg *Config) {
	if cfg.App.Port == 0 {
		cfg.App.Port = 8080
	}
	if cfg.App.Name == "" {
		cfg.App.Name = "Sekolah App"
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "./data/sekolah.db"
	}
	if cfg.Upload.MaxSize == 0 {
		cfg.Upload.MaxSize = 5
	}
	if cfg.Backup.Retention == 0 {
		cfg.Backup.Retention = 7
	}
	if cfg.Backup.Path == "" {
		cfg.Backup.Path = "./backups"
	}
	if cfg.RateLimit.Login == 0 {
		cfg.RateLimit.Login = 5
	}
	if cfg.RateLimit.Register == 0 {
		cfg.RateLimit.Register = 3
	}
	if cfg.RateLimit.API == 0 {
		cfg.RateLimit.API = 100
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
	if cfg.Logging.Output == "" {
		cfg.Logging.Output = "stdout"
	}
}

func validate(cfg *Config) error {
	if cfg.App.Port < 1 || cfg.App.Port > 65535 {
		return fmt.Errorf("app.port must be between 1 and 65535, got %d", cfg.App.Port)
	}
	if cfg.Upload.MaxSize < 1 {
		return fmt.Errorf("upload.max_size must be positive")
	}
	return nil
}

func ValidateSecrets(secrets *Secrets) error {
	if len(secrets.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	return nil
}
