# Konfigurasi

## File Konfigurasi

Aplikasi menggunakan dua file konfigurasi:
- `config.yaml` — Pengaturan aplikasi
- `.env` — Secrets dan API keys

---

## config.yaml

```yaml
# Pengaturan aplikasi
app:
  port: 8080                          # Port aplikasi
  name: "Sekolah App"                 # Nama aplikasi
  base_url: "http://localhost:8080"   # Base URL

# Database
database:
  path: "./data/sekolah.db"           # Path database SQLite

# Modul aktif
modules:
  ppdb: true                          # Modul PPDB
  pembayaran: true                    # Modul Pembayaran
  notifikasi: true                    # Modul Notifikasi

# Notifikasi
notifikasi:
  whatsapp: true                      # Aktifkan WhatsApp
  telegram: false                     # Aktifkan Telegram
  email: true                         # Aktifkan Email

# Upload
upload:
  max_size: 5                         # Max upload (MB)
  allowed_types:                      # Tipe file yang diizinkan
    - image/jpeg
    - image/png
    - application/pdf

# Backup
backup:
  enabled: true                       # Aktifkan auto backup
  schedule: "0 2 * * *"              # Cron schedule (jam 2 pagi)
  retention: 7                        # Simpan backup (hari)
  path: "./backups"                   # Path backup

# Rate limiting
rate_limit:
  login: 5                            # Login per menit
  register: 3                         # Register per menit
  api: 100                            # API per menit

# CORS
cors:
  allowed_origins:                    # Origin yang diizinkan
    - "http://localhost:5173"
    - "http://localhost:8080"

# Logging
logging:
  level: "info"                       # debug, info, warn, error
  format: "json"                      # json, text
  output: "stdout"                    # stdout, file
  file: "./logs/app.log"             # Path log file
```

---

## .env

```env
# JWT
JWT_SECRET=your-secret-key-min-32-characters

# Google OAuth (opsional)
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=

# SMTP Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# Telegram Bot (opsional)
TELEGRAM_BOT_TOKEN=

# Payment Gateway (opsional)
MIDTRANS_SERVER_KEY=
MIDTRANS_CLIENT_KEY=
XENDIT_SECRET_KEY=
```

---

## Environment Variables

Selain file `.env`, bisa juga pakai environment variables:

| Variable | Keterangan | Default |
|----------|------------|---------|
| `PORT` | Port aplikasi | 8080 |
| `DB_PATH` | Path database | ./data/sekolah.db |
| `JWT_SECRET` | Secret key JWT | (wajib) |
| `LOG_LEVEL` | Level logging | info |

---

## Konfigurasi WhatsApp

### Setup whatsmeow

1. Jalankan aplikasi
2. Buka halaman Pengaturan > Notifikasi > WhatsApp
3. Scan QR code dengan WhatsApp
4. Tunggu hingga status "Terhubung"

### Catatan
- WhatsApp akan reconnect otomatis jika terputus
- Jika ganti nomor, perlu scan ulang
- Database session tersimpan di `data/whatsapp.db`

---

## Konfigurasi Telegram

### Setup Bot

1. Chat @BotFather di Telegram
2. Buat bot baru dengan `/newbot`
3. Dapatkan bot token
4. Masukkan token ke `.env`

```env
TELEGRAM_BOT_TOKEN=123456789:ABCdefGHIjklMNOpqrsTUVwxyz
```

### Setup Chat

1. Chat bot yang sudah dibuat
2. Dapatkan chat ID dari `https://api.telegram.org/bot<TOKEN>/getUpdates`
3. Masukkan chat ID di pengaturan aplikasi

---

## Konfigurasi Email (SMTP)

### Gmail

1. Aktifkan 2FA di Google Account
2. Buat App Password di Security > App Passwords
3. Masukkan ke `.env`

```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-16-char-app-password
```

### SMTP Lain

```env
SMTP_HOST=smtp.your-provider.com
SMTP_PORT=587
SMTP_USER=your-email@domain.com
SMTP_PASSWORD=your-password
```

---

## Konfigurasi Payment Gateway

### Midtrans

1. Daftar di https://midtrans.com
2. Dapatkan Server Key dan Client Key
3. Masukkan ke `.env`

```env
MIDTRANS_SERVER_KEY=SB-Mid-server-xxx
MIDTRANS_CLIENT_KEY=SB-Mid-client-xxx
```

4. Set callback URL di Midtrans Dashboard:
   ```
   https://your-domain.com/api/v1/payment/callback/midtrans
   ```

### Xendit

1. Daftar di https://xendit.co
2. Dapatkan Secret Key
3. Masukkan ke `.env`

```env
XENDIT_SECRET_KEY=xnd_xxx
```

4. Set callback URL di Xendit Dashboard:
   ```
   https://your-domain.com/api/v1/payment/callback/xendit
   ```

---

## Konfigurasi Google OAuth

1. Buka https://console.cloud.google.com
2. Buat project baru
3. Aktifkan Google+ API
4. Buat OAuth 2.0 credentials
5. Set authorized redirect URI:
   ```
   http://localhost:8080/api/v1/auth/google/callback
   ```
6. Masukkan Client ID dan Secret ke `.env`

```env
GOOGLE_CLIENT_ID=xxx.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=xxx
```
