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
SMTP_FROM=noreply@sekolah.com  # opsional, fallback ke SMTP_USER

# Telegram Bot (opsional)
TELEGRAM_BOT_TOKEN=
TELEGRAM_BOT_USERNAME=YourBotUsername
TELEGRAM_WEBHOOK_SECRET=random-secret-string

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

## Validasi Konfigurasi

Aplikasi memvalidasi semua konfigurasi saat startup. Jika ada konfigurasi yang invalid atau wajib tapi kosong, aplikasi akan gagal start dengan pesan error yang jelas.

### Validasi Wajib

| Konfigurasi | Validasi | Aksi Jika Invalid |
|-------------|----------|-------------------|
| `JWT_SECRET` | Minimal 32 karakter | Gagal start |
| `database.path` | Path harus writable | Gagal start |
| `app.port` | Angka 1-65535 | Gagal start |
| `backup.schedule` | Format cron valid | Warning, backup dinonaktifkan |
| `upload.max_size` | Angka positif | Default ke 5 MB |
| `rate_limit.*` | Angka positif | Default ke nilai bawaan |

### Validasi Opsional

Konfigurasi opsional hanya divalidasi jika modul terkait aktif:

| Modul | Konfigurasi | Validasi |
|-------|-------------|----------|
| Notifikasi WhatsApp | `notifikasi.whatsapp: true` | Cek file session `data/whatsapp.db` |
| Notifikasi Telegram | `TELEGRAM_BOT_TOKEN` | Format token valid |
| Notifikasi Email | `SMTP_HOST`, `SMTP_PORT` | Wajib jika email aktif |
| Payment Midtrans | `MIDTRANS_SERVER_KEY` | Wajib jika midtrans aktif |
| Payment Xendit | `XENDIT_SECRET_KEY` | Wajib jika xendit aktif |
| Google OAuth | `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` | Wajib jika Google login aktif |

### Perilaku Startup

```
1. Load config.yaml
2. Load .env (override environment variables)
3. Validasi konfigurasi wajib
   → Jika gagal: log error + exit code 1
4. Validasi konfigurasi opsional (per modul aktif)
   → Jika gagal: log warning + nonaktifkan modul terkait
5. Tampilkan ringkasan konfigurasi di log:
   - Port
   - Modul aktif
   - Notifikasi aktif
   - Payment gateway aktif
6. Lanjut ke migration & serve
```

### Contoh Output Startup

```
INFO  SekolahApp v1.2.0 starting...
INFO  Config loaded from ./config.yaml
INFO  Database: ./data/sekolah.db
INFO  Modules: pembayaran=ON, ppdb=ON, notifikasi=ON
INFO  Notifikasi: whatsapp=ON, telegram=OFF, email=ON
WARN  Telegram dinonaktifkan: TELEGRAM_BOT_TOKEN kosong
INFO  Payment: midtrans=OFF, xendit=OFF
INFO  Backup: ON (schedule: 0 2 * * *, retention: 7 hari)
INFO  Server listening on :8080
```

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

### Opt-in Flow

Telegram bot tidak bisa memulai chat. Penerima harus:
1. Admin buat link undangan dari halaman Preferensi
2. Kirim link ke penerima (orang tua/siswa)
3. Penerima klik link → bot Telegram terbuka → kirim /start
4. Bot mengikat chat_id ke preferensi → consent otomatis jadi `granted`

Setup webhook:
```
POST https://your-domain.com/api/v1/telegram/webhook
```
Set webhook secret di Telegram Bot API:
```
https://api.telegram.org/bot<TOKEN>/setWebhook?url=<WEBHOOK_URL>&secret_token=<SECRET>
```

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
SMTP_FROM=noreply@sekolah.com  # opsional
```

### SMTP Lain

```env
SMTP_HOST=smtp.your-provider.com
SMTP_PORT=587
SMTP_USER=your-email@domain.com
SMTP_PASSWORD=your-password
```

### Consent & Preferensi

Worker tidak mengirim notifikasi jika:
- Tidak ada preferensi untuk kanal+tujuan tersebut
- `enabled = false`
- `consent_status != 'granted'`

Status konsen: `pending` | `granted` | `revoked`

Kelola via admin: `GET/PUT /api/v1/notifikasi/preferensi`

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
