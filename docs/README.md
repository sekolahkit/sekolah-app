# SekolahApp - Dokumentasi

Sistem pembayaran sekolah dan PPDB (Penerimaan Peserta Didik Baru) berbasis web, open-source, untuk sekolah di Indonesia.

## Fitur Utama

- **Pembayaran** — SPP, uang pangkal, dan pembayaran lainnya dengan support cicilan
- **PPDB** — Pendaftaran siswa baru secara online dengan modul modular
- **Notifikasi** — WhatsApp (whatsmeow), Telegram, dan Email
- **Laporan** — Dashboard, export Excel, dan cetak kwitansi
- **Multi-role** — Admin, Operator, Guru, Siswa, Orangtua

## Dokumentasi

| Dokumen | Deskripsi |
|---------|-----------|
| [Arsitektur](architecture.md) | Stack teknologi, struktur folder, keputusan desain |
| [Database](database.md) | Schema database, tabel, relasi |
| [API](api.md) | Struktur REST API, endpoint, versioning |
| [Modul](modules.md) | Detail modul: Pembayaran, PPDB, Siswa, Kelas, Notifikasi |
| [Setup](setup.md) | Instalasi dan konfigurasi awal |
| [Konfigurasi](configuration.md) | File konfigurasi, environment variables |
| [Keamanan](security.md) | Autentikasi, otorisasi, keamanan |
| [Deployment](deployment.md) | Cara deploy ke server |
| [Contributing](contributing.md) | Cara berkontribusi ke project |

## Stack Teknologi

| Layer | Teknologi |
|-------|-----------|
| Backend | Go + Chi + Squirrel |
| Frontend | React + Vite + shadcn/ui + TanStack |
| Database | SQLite |
| Notifikasi | whatsmeow (WhatsApp), Telegram Bot API, SMTP (Email) |
| Payment Gateway | Adapter pattern (Midtrans, Xendit, dll) |

## Lisensi

MIT License
