# SekolahApp

Sistem pembayaran sekolah dan PPDB (Penerimaan Peserta Didik Baru) berbasis web, open-source, untuk sekolah di Indonesia.

## Fitur

- **Pembayaran** — SPP, uang pangkal, cicilan, verifikasi manual & payment gateway (Midtrans, Xendit)
- **PPDB** — Pendaftaran online, verifikasi berkas, pengumuman, daftar ulang
- **Notifikasi** — WhatsApp, Telegram, Email
- **Laporan** — Dashboard, rekap pembayaran/PPDB/siswa, ekspor Excel
- **Multi-role** — Admin, Operator, Guru, Siswa, Orangtua
- **Multi-tenant** — Satu instance untuk banyak sekolah

## Tech Stack

| Layer | Teknologi |
|-------|-----------|
| Backend | Go, Chi router, Squirrel query builder |
| Frontend | React 19, Vite, TailwindCSS, TanStack Query |
| Database | SQLite (WAL mode) |
| Notifikasi | whatsmeow, Telegram Bot API, SMTP |
| Payment | Midtrans, Xendit (adapter pattern) |

## Quick Start

```bash
# Clone
git clone https://github.com/sekolahkit/sekolah-app.git
cd sekolah-app

# Backend
cd backend
cp .env.example .env
# Edit .env dengan JWT_SECRET minimal 32 karakter
make dev

# Frontend (terminal terpisah)
cd frontend
npm install
npm run dev
```

Buka `http://localhost:5173` dan ikuti setup wizard.

## Dev Setup

**Prasyarat:**
- Go 1.23+
- Node.js 22+
- npm 10+

```bash
# Install dependencies
cd frontend && npm install

# Jalankan backend
cd backend && make dev

# Jalankan frontend dev server
cd frontend && npm run dev

# Jalankan test
cd backend && make test
cd frontend && npm run lint
```

## Production Build

```bash
cd backend && make build
```

Perintah ini akan:
1. Build frontend (`npm run build`)
2. Embed hasil build ke dalam binary Go
3. Menghasilkan binary `sekolah-app` yang siap deploy

Jalankan binary:
```bash
./sekolah-app
```

## Konfigurasi

Lihat [docs/configuration.md](docs/configuration.md) untuk referensi lengkap `config.yaml` dan environment variables.

## Dokumentasi

- [Arsitektur](docs/architecture.md)
- [API Reference](docs/api.md)
- [Database Schema](docs/database.md)
- [Modul](docs/modules.md)
- [Setup & Instalasi](docs/setup.md)
- [Konfigurasi](docs/configuration.md)
- [Deployment](docs/deployment.md)
- [Keamanan](docs/security.md)
- [Kontribusi](docs/contributing.md)

## Dukung Pengembangan

SekolahApp dikembangkan sebagai proyek open-source untuk membantu sekolah mengelola pembayaran, PPDB, notifikasi, dan laporan.

Jika proyek ini bermanfaat, Anda bisa mendukung pengembangan melalui [Saweria](https://saweria.co/erwisnu).

Daftar pendukung bisa dilihat di [docs/supporters.md](docs/supporters.md).

## Lisensi

[MIT](LICENSE)
