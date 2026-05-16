# API Reference

## Overview

API menggunakan REST dengan format JSON. Semua endpoint di-prefix dengan `/api/v1/`.

## Base URL

```
http://localhost:8080/api/v1
```

## Versioning Strategy

API menggunakan URL-based versioning (`/api/v1/`, `/api/v2/`).

| Kebijakan | Keterangan |
|-----------|------------|
| Backward compatible | Penambahan field baru di response tidak dianggap breaking change |
| Deprecation notice | Versi lama akan diberi header `X-API-Deprecated: true` minimal 6 bulan sebelum dihapus |
| Breaking changes | Perubahan yang menghapus/mengubah field existing → versi baru |
| Support window | Maksimal 2 versi aktif bersamaan |

Header response untuk versi yang deprecated:
```
X-API-Deprecated: true
X-API-Sunset: 2025-06-01
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
| POST | `/auth/logout` | Logout, revoke refresh token | Auth |
| POST | `/auth/refresh` | Refresh access token (baca refresh_token dari cookie) | Public (cookie required) |
| POST | `/auth/revoke-all/:user_id` | Revoke semua session user | Admin |
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

### Rekening Sekolah

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| GET | `/rekening-sekolah` | List semua rekening | Admin, Operator |
| GET | `/rekening-sekolah/aktif` | List rekening aktif (tampil di halaman bayar) | Auth |
| POST | `/rekening-sekolah` | Tambah rekening | Admin |
| PUT | `/rekening-sekolah/:id` | Update rekening | Admin |
| DELETE | `/rekening-sekolah/:id` | Nonaktifkan rekening (soft delete: aktif=false) | Admin |

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
| POST | `/payment/callback/midtrans` | Callback Midtrans (validasi SHA-512 signature) | Public |
| POST | `/payment/callback/xendit` | Callback Xendit (validasi x-callback-token header) | Public |

> **Catatan:** Endpoint callback tidak memerlukan auth cookie, tapi divalidasi via signature/token dari masing-masing provider. Callback bersifat idempotent — request dengan `payment_gateway_id` yang sudah diproses akan return 200 tanpa side effect.

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
| GET | `/setup/status` | Cek status setup | Public (hanya aktif jika belum initialized) |
| POST | `/setup` | Jalankan setup wizard | Public (hanya aktif jika belum initialized) |

> **Setup Guard:** Kedua endpoint ini hanya aktif selama database belum initialized (tabel `sekolah` kosong). Setelah setup selesai, endpoint akan return `404 Not Found`. Ini mencegah takeover atau re-setup setelah deploy.

### Upload

| Method | Endpoint | Keterangan | Role |
|--------|----------|------------|------|
| POST | `/upload` | Upload file | Auth |
| GET | `/upload/:path` | Download file (cek auth + relasi) | Auth |
| GET | `/upload/signed/:token` | Download file via signed URL (TTL 5-15 menit) | Public |

> **File Access Control:**
> - Semua file privat (bukti bayar, KK, akta, berkas PPDB, foto siswa, dokumen pembayaran) hanya bisa diakses lewat endpoint auth handler.
> - Handler cek: (1) user authenticated, (2) role diizinkan, (3) user punya relasi ke data terkait (via tenant scoping + pengguna_siswa).
> - **Signed URL** digunakan untuk kasus khusus: link sementara di notifikasi WhatsApp/Telegram/Email, preview file, atau sharing terbatas. TTL pendek (5-15 menit), single-use opsional.
> - Public assets (logo sekolah) disimpan terpisah di folder `/public/` dan boleh di-serve langsung tanpa auth.

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

---

## Contoh Request & Response

### POST /auth/login

**Request:**
```json
{
    "kode_sekolah": "sdn1bdg",
    "email": "admin@sekolah.id",
    "password": "password123"
}
```

**Response (200):**
```json
{
    "data": {
        "id": 1,
        "nama": "Admin Sekolah",
        "email": "admin@sekolah.id",
        "role": "admin",
        "sekolah_id": 1
    }
}
```

Cookies yang di-set:
```
Set-Cookie: access_token=eyJhbG...; HttpOnly; Secure; SameSite=Strict; Path=/; Max-Age=900
Set-Cookie: refresh_token=rft_abc...; HttpOnly; Secure; SameSite=Strict; Path=/api/v1/auth/refresh; Max-Age=604800
```

---

### POST /auth/refresh

Refresh access token. Refresh token dibaca dari cookie (bukan body).

**Request:** (tidak ada body, refresh_token dikirim otomatis via cookie)

**Response (200):**
```json
{
    "data": {
        "id": 1,
        "nama": "Admin Sekolah",
        "email": "admin@sekolah.id",
        "role": "admin",
        "sekolah_id": 1
    }
}
```

Cookies yang di-set (rotation):
```
Set-Cookie: access_token=eyJhbG...(baru); HttpOnly; Secure; SameSite=Strict; Path=/; Max-Age=900
Set-Cookie: refresh_token=rft_def...(baru); HttpOnly; Secure; SameSite=Strict; Path=/api/v1/auth/refresh; Max-Age=604800
```

> **CSRF:** Endpoint ini aman dari CSRF karena `SameSite=Strict` + path cookie dibatasi ke `/api/v1/auth/refresh`. Tidak ada state-changing side effect selain rotate token.

---

### POST /siswa

**Request:**
```json
{
    "nis": "2024001",
    "nama": "Andi Pratama",
    "jenis_kelamin": "L",
    "tempat_lahir": "Jakarta",
    "tanggal_lahir": "2010-05-15",
    "agama": "Islam",
    "alamat": "Jl. Merdeka No. 10",
    "no_hp": "081234567890",
    "nama_ortu": "Budi Pratama",
    "no_hp_ortu": "081298765432"
}
```

**Response (201):**
```json
{
    "data": {
        "id": 42,
        "nis": "2024001",
        "nama": "Andi Pratama",
        "jenis_kelamin": "L",
        "status": "aktif",
        "created_at": "2024-07-01T10:00:00Z"
    }
}
```

---

### POST /tagihan/bulk

Buat tagihan massal untuk banyak siswa sekaligus.

**Request:**
```json
{
    "siswa_ids": [1, 2, 3, 4, 5],
    "kategori_id": 1,
    "tahun_ajaran_id": 2,
    "semester": "Ganjil",
    "nominal": 500000,
    "jatuh_tempo": "2024-08-15",
    "catatan": "SPP Bulan Agustus"
}
```

**Response (201):**
```json
{
    "data": {
        "created_count": 5,
        "tagihan_ids": [101, 102, 103, 104, 105]
    }
}
```

---

### POST /pembayaran

Upload bukti bayar oleh siswa/orangtua.

**Request (multipart/form-data):**
```
tagihan_id: 101
jumlah: 250000
metode: transfer
rekening_sekolah_id: 2
bukti_bayar: [file upload]
catatan: "Cicilan pertama"
```

> `rekening_sekolah_id` wajib jika metode=transfer. NULL/tidak dikirim jika metode=cash atau gateway.

**Response (201):**
```json
{
    "data": {
        "id": 55,
        "tagihan_id": 101,
        "jumlah": 250000,
        "metode": "transfer",
        "status": "pending",
        "created_at": "2024-08-10T14:30:00Z"
    }
}
```

---

### PUT /pembayaran/:id/verify

**Request:**
```json
{
    "catatan": "Bukti transfer valid"
}
```

**Response (200):**
```json
{
    "data": {
        "id": 55,
        "status": "verified",
        "verified_by": 1,
        "verified_at": "2024-08-10T15:00:00Z"
    }
}
```

---

### POST /ppdb/daftar

**Request (multipart/form-data):**
```
nama_lengkap: "Siti Nurhaliza"
nik: "3201234567890001"
tempat_lahir: "Bandung"
tanggal_lahir: "2011-03-20"
jenis_kelamin: "P"
agama: "Islam"
alamat: "Jl. Asia Afrika No. 5"
asal_sekolah: "SD Negeri 1 Bandung"
no_hp: "081300001111"
email: "siti@email.com"
nama_ortu: "Ahmad Nurhaliza"
no_hp_ortu: "081300002222"
pekerjaan_ortu: "Wiraswasta"
foto: [file upload]
berkas[]: [file upload - ijazah]
berkas[]: [file upload - akta_lahir]
berkas[]: [file upload - kk]
```

**Response (201):**
```json
{
    "data": {
        "id": 15,
        "nama_lengkap": "Siti Nurhaliza",
        "status": "menunggu",
        "created_at": "2024-06-01T09:00:00Z"
    }
}
```

---

### POST /payment/callback/midtrans

Callback dari Midtrans (server-to-server).

**Request (dari Midtrans):**
```json
{
    "transaction_id": "trx-001",
    "order_id": "pembayaran-55",
    "transaction_status": "settlement",
    "gross_amount": "250000.00",
    "payment_type": "bank_transfer",
    "signature_key": "abc123..."
}
```

**Response (200):**
```json
{
    "status": "ok"
}
```

---

### POST /siswa/import

Import siswa dari file Excel.

**Request (multipart/form-data):**
```
file: [file upload - .xlsx]
```

**Response (200):**
```json
{
    "data": {
        "total": 50,
        "berhasil": 48,
        "gagal": 2,
        "errors": [
            {"baris": 12, "pesan": "NIS sudah terdaftar: 2024005"},
            {"baris": 35, "pesan": "Nama wajib diisi"}
        ]
    }
}
```

---

### POST /rekening-sekolah

**Request:**
```json
{
    "nama_bank": "BCA",
    "nomor_rekening": "1234567890",
    "nama_pemilik": "SDN 1 Bandung",
    "cabang": "KCP Dago",
    "catatan": "Rekening utama SPP"
}
```

**Response (201):**
```json
{
    "data": {
        "id": 2,
        "nama_bank": "BCA",
        "nomor_rekening": "1234567890",
        "nama_pemilik": "SDN 1 Bandung",
        "cabang": "KCP Dago",
        "aktif": true,
        "urutan": 0,
        "created_at": "2024-07-01T10:00:00Z"
    }
}
```

---

### GET /rekening-sekolah/aktif

List rekening aktif untuk ditampilkan di halaman pembayaran siswa/orangtua.

**Response (200):**
```json
{
    "data": [
        {
            "id": 1,
            "nama_bank": "BRI",
            "nomor_rekening": "0987654321",
            "nama_pemilik": "SDN 1 Bandung",
            "cabang": "KC Bandung"
        },
        {
            "id": 2,
            "nama_bank": "BCA",
            "nomor_rekening": "1234567890",
            "nama_pemilik": "SDN 1 Bandung",
            "cabang": "KCP Dago"
        }
    ]
}
```
