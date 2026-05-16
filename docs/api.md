# API Reference

## Overview

API menggunakan REST dengan format JSON. Semua endpoint di-prefix dengan `/api/v1/`.

## Base URL

```
http://localhost:8080/api/v1
```

## Autentikasi

Sebagian besar endpoint membutuhkan JWT token. Token dikirim via httpOnly cookie.

### Public Endpoints (Tanpa Auth)
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/google`
- `GET /api/v1/ppdb/daftar` (halaman publik PPDB)
- `POST /api/v1/ppdb/daftar` (submit pendaftaran)
- `GET /api/v1/ppdb/pengumuman/:id`

## Format Response

### Success
```json
{
    "data": { ... },
    "meta": {
        "page": 1,
        "limit": 20,
        "total": 100,
        "total_pages": 5
    }
}
```

### Error
```json
{
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Nama wajib diisi",
        "details": {
            "field": "nama"
        }
    }
}
```

## Pagination

Semua endpoint list mendukung pagination:

```
GET /api/v1/siswa?page=1&limit=20&sort=nama&search=andi
```

| Parameter | Keterangan | Default |
|-----------|------------|---------|
| `page` | Halaman ke- | 1 |
| `limit` | Jumlah data per halaman | 20 |
| `sort` | Field untuk sorting | `created_at` |
| `order` | `asc` atau `desc` | `desc` |
| `search` | Keyword pencarian | - |

## Endpoints

### Auth

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| POST | `/auth/login` | Login email/password | Public |
| POST | `/auth/register` | Register user baru | Public |
| POST | `/auth/google` | Login dengan Google | Public |
| POST | `/auth/logout` | Logout | Auth |
| GET | `/auth/me` | Data user saat ini | Auth |
| PUT | `/auth/password` | Ganti password | Auth |

### Sekolah

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| GET | `/sekolah` | Data sekolah | Auth |
| PUT | `/sekolah` | Update sekolah | Admin |

### Users

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| GET | `/users` | List semua user | Admin |
| GET | `/users/:id` | Detail user | Admin |
| POST | `/users` | Buat user baru | Admin |
| PUT | `/users/:id` | Update user | Admin |
| DELETE | `/users/:id` | Hapus user | Admin |

### Siswa

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| GET | `/siswa` | List siswa | Admin, Operator |
| GET | `/siswa/:id` | Detail siswa | Admin, Operator |
| POST | `/siswa` | Tambah siswa | Admin, Operator |
| PUT | `/siswa/:id` | Update siswa | Admin, Operator |
| DELETE | `/siswa/:id` | Hapus siswa | Admin |
| POST | `/siswa/import` | Import dari Excel | Admin, Operator |
| GET | `/siswa/export` | Export ke Excel | Admin, Operator |
| GET | `/siswa/:id/tagihan` | Tagihan siswa | Admin, Operator, Siswa, Orangtua |
| GET | `/siswa/:id/pembayaran` | Riwayat pembayaran | Admin, Operator, Siswa, Orangtua |

### Kelas

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| GET | `/kelas` | List kelas | Admin, Operator |
| GET | `/kelas/:id` | Detail kelas | Admin, Operator |
| POST | `/kelas` | Tambah kelas | Admin |
| PUT | `/kelas/:id` | Update kelas | Admin |
| DELETE | `/kelas/:id` | Hapus kelas | Admin |
| GET | `/kelas/:id/siswa` | Siswa di kelas | Admin, Operator, Guru |
| POST | `/kelas/:id/siswa` | Tambah siswa ke kelas | Admin |
| DELETE | `/kelas/:id/siswa/:siswa_id` | Hapus siswa dari kelas | Admin |

### Jurusan

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| GET | `/jurusan` | List jurusan | Admin |
| POST | `/jurusan` | Tambah jurusan | Admin |
| PUT | `/jurusan/:id` | Update jurusan | Admin |
| DELETE | `/jurusan/:id` | Hapus jurusan | Admin |

### Tahun Ajaran

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| GET | `/tahun-ajaran` | List tahun ajaran | Admin, Operator |
| GET | `/tahun-ajaran/aktif` | Tahun ajaran aktif | Auth |
| POST | `/tahun-ajaran` | Tambah tahun ajaran | Admin |
| PUT | `/tahun-ajaran/:id` | Update tahun ajaran | Admin |
| PUT | `/tahun-ajaran/:id/aktif` | Set sebagai aktif | Admin |

### Kategori Pembayaran

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| GET | `/kategori-pembayaran` | List kategori | Admin, Operator |
| POST | `/kategori-pembayaran` | Tambah kategori | Admin |
| PUT | `/kategori-pembayaran/:id` | Update kategori | Admin |
| DELETE | `/kategori-pembayaran/:id` | Hapus kategori | Admin |

### Tagihan

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| GET | `/tagihan` | List tagihan | Admin, Operator |
| GET | `/tagihan/:id` | Detail tagihan | Admin, Operator, Siswa, Orangtua |
| POST | `/tagihan` | Buat tagihan | Admin, Operator |
| PUT | `/tagihan/:id` | Update tagihan | Admin, Operator |
| DELETE | `/tagihan/:id` | Hapus tagihan | Admin |
| POST | `/tagihan/bulk` | Buat tagihan massal | Admin, Operator |

### Pembayaran

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| GET | `/pembayaran` | List pembayaran | Admin, Operator |
| GET | `/pembayaran/:id` | Detail pembayaran | Admin, Operator |
| POST | `/pembayaran` | Upload bukti bayar | Siswa, Orangtua |
| PUT | `/pembayaran/:id/verify` | Verifikasi pembayaran | Admin, Operator |
| PUT | `/pembayaran/:id/reject` | Tolak pembayaran | Admin, Operator |
| GET | `/pembayaran/export` | Export ke Excel | Admin, Operator |

### Payment Gateway

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| POST | `/payment/callback/midtrans` | Callback Midtrans | Public |
| POST | `/payment/callback/xendit` | Callback Xendit | Public |

### PPDB (Publik)

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| GET | `/ppdb/daftar` | Form pendaftaran | Public |
| POST | `/ppdb/daftar` | Submit pendaftaran | Public |
| GET | `/ppdb/pengumuman/:id` | Cek pengumuman | Public |

### PPDB (Admin)

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| GET | `/ppdb/pendaftar` | List pendaftar | Admin, Operator |
| GET | `/ppdb/pendaftar/:id` | Detail pendaftar | Admin, Operator |
| PUT | `/ppdb/pendaftar/:id` | Update status | Admin, Operator |
| GET | `/ppdb/pendaftar/:id/berkas` | List berkas | Admin, Operator |
| PUT | `/ppdb/berkas/:id` | Verifikasi berkas | Admin, Operator |
| POST | `/ppdb/ujian` | Input nilai ujian | Admin, Operator |
| POST | `/ppdb/pengumuman` | Publish pengumuman | Admin |
| GET | `/ppdb/export` | Export ke Excel | Admin, Operator |

### Notifikasi

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| GET | `/notifikasi` | List notifikasi terkirim | Admin |
| POST | `/notifikasi/test` | Test kirim notifikasi | Admin |
| GET | `/notifikasi/queue` | Status antrian | Admin |

### Laporan

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| GET | `/laporan/pembayaran` | Rekap pembayaran | Admin, Operator |
| GET | `/laporan/pembayaran/export` | Export rekap Excel | Admin, Operator |
| GET | `/laporan/ppdb` | Rekap PPDB | Admin, Operator |
| GET | `/laporan/ppdb/export` | Export rekap Excel | Admin, Operator |
| GET | `/laporan/siswa` | Rekap siswa | Admin, Operator |
| GET | `/laporan/siswa/export` | Export rekap Excel | Admin, Operator |

### Dashboard

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| GET | `/dashboard/admin` | Statistik admin | Admin |
| GET | `/dashboard/operator` | Statistik operator | Operator |
| GET | `/dashboard/guru` | Statistik guru | Guru |
| GET | `/dashboard/siswa` | Statistik siswa | Siswa |
| GET | `/dashboard/orangtua` | Statistik orangtua | Orangtua |

### Backup

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| GET | `/backup` | List backup | Admin |
| POST | `/backup` | Buat backup manual | Admin |
| GET | `/backup/:id/download` | Download backup | Admin |
| POST | `/backup/restore/:id` | Restore backup | Admin |

### Setup

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| GET | `/setup/status` | Cek status setup | Public |
| POST | `/setup` | Jalankan setup wizard | Public |

### Upload

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| POST | `/upload` | Upload file | Auth |
| GET | `/upload/:path` | Download file | Auth |

## HTTP Status Codes

| Code | Keterangan |
|------|------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request (validation error) |
| 401 | Unauthorized (belum login) |
| 403 | Forbidden (tidak punya akses) |
| 404 | Not Found |
| 422 | Unprocessable Entity (business logic error) |
| 429 | Too Many Requests (rate limit) |
| 500 | Internal Server Error |

## Rate Limiting

| Endpoint | Limit |
|----------|-------|
| `POST /auth/login` | 5 request/menit |
| `POST /auth/register` | 3 request/menit |
| `POST /ppdb/daftar` | 10 request/menit |
| Lainnya | 100 request/menit |
