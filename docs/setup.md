# Instalasi & Setup

## Persyaratan

### Minimum
- OS: Linux, Windows, atau macOS
- RAM: 512 MB
- Storage: 100 MB (belum termasuk upload)
- Port: 8080 (bisa dikonfigurasi)

### Untuk Development
- Go 1.22+
- Node.js 18+
- npm atau yarn

---

## Instalasi Binary

### 1. Download Binary

Download dari GitHub Releases:

```bash
# Linux
wget https://github.com/Sekolahkit/sekolah-app/releases/latest/download/sekolah-app-linux-amd64
chmod +x sekolah-app-linux-amd64

# Windows
# Download sekolah-app-windows.exe

# macOS
wget https://github.com/Sekolahkit/sekolah-app/releases/latest/download/sekolah-app-darwin-amd64
chmod +x sekolah-app-darwin-amd64
```

### 2. Jalankan

```bash
./sekolah-app-linux-amd64
```

Aplikasi akan berjalan di `http://localhost:8080`.

### 3. Setup Awal

Buka browser dan akses `http://localhost:8080/setup`:

1. **Selamat Datang** — Pengenalan aplikasi
2. **Info Sekolah** — Nama, alamat, logo sekolah
3. **Akun Admin** — Buat akun administrator
4. **Modul Aktif** — Pilih modul yang digunakan
5. **Notifikasi** — Konfigurasi WhatsApp/Telegram/Email
6. **Selesai** — Aplikasi siap digunakan

---

## Instalasi Docker

### 1. Buat docker-compose.yml

```yaml
version: '3.8'

services:
  sekolah-app:
    image: sekolahkit/sekolah-app:latest
    container_name: sekolah-app
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
      - ./uploads:/app/uploads
      - ./backups:/app/backups
      - ./config.yaml:/app/config.yaml
      - ./.env:/app/.env
    restart: unless-stopped
```

### 2. Buat config.yaml

```yaml
app:
  port: 8080
  name: "Sekolah App"

database:
  path: "./data/sekolah.db"

modules:
  ppdb: true
  pembayaran: true
  notifikasi: true

notifikasi:
  whatsapp: true
  telegram: false
  email: true

upload:
  max_size: 5
  allowed_types:
    - image/jpeg
    - image/png
    - application/pdf

backup:
  enabled: true
  schedule: "0 2 * * *"
  retention: 7
  path: "./backups"
```

### 3. Buat .env

```env
JWT_SECRET=your-secret-key-here
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=
TELEGRAM_BOT_TOKEN=
MIDTRANS_SERVER_KEY=
XENDIT_SECRET_KEY=
```

### 4. Jalankan

```bash
docker-compose up -d
```

---

## Development Setup

### 1. Clone Repository

```bash
git clone https://github.com/Sekolahkit/sekolah-app.git
cd sekolah-app
```

### 2. Backend

```bash
cd backend

# Install dependencies
go mod tidy

# Jalankan
go run cmd/server/main.go
```

### 3. Frontend

```bash
cd frontend

# Install dependencies
npm install

# Jalankan development server
npm run dev
```

Frontend akan berjalan di `http://localhost:5173` dengan proxy ke backend di `http://localhost:8080`.

### 4. Build Frontend untuk Production

```bash
cd frontend
npm run build
```

Output akan di-embed ke binary Go saat build.

### 5. Build Binary

```bash
cd backend

# Build untuk Linux
GOOS=linux GOARCH=amd64 go build -o sekolah-app cmd/server/main.go

# Build untuk Windows
GOOS=windows GOARCH=amd64 go build -o sekolah-app.exe cmd/server/main.go

# Build untuk macOS
GOOS=darwin GOARCH=amd64 go build -o sekolah-app-darwin cmd/server/main.go
```

---

## Struktur File Setelah Install

```
sekolah-app/
├── sekolah-app          # Binary
├── config.yaml          # Konfigurasi
├── .env                 # Secrets
├── data/
│   └── sekolah.db       # Database SQLite
├── uploads/
│   /1/                  # Upload per sekolah_id
│   │   /bukti_bayar/
│   │   /berkas_ppdb/
│   │   /foto_siswa/
├── backups/
│   /2024-01-15/
│   │   sekolah.db
│   │   uploads.tar.gz
└── logs/
    └── app.log
```

---

## Troubleshooting

### Port sudah digunakan
```bash
# Cek proses di port 8080
lsof -i :8080

# Atau gunakan port lain
./sekolah-app --port 9090
```

### Permission denied
```bash
chmod +x sekolah-app-linux-amd64
```

### Database locked
Pastikan tidak ada proses lain yang mengakses database.

### WhatsApp tidak terhubung
1. Pastikan nomor WhatsApp aktif
2. Scan QR code yang muncul di terminal
3. Pastikan koneksi internet stabil
