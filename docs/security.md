# Keamanan

## Overview

Aplikasi menggunakan beberapa lapisan keamanan untuk melindungi data sekolah.

---

## Autentikasi

### Session Strategy (Dual Cookie)

Aplikasi menggunakan dua httpOnly cookie untuk autentikasi:

| Cookie | TTL | Path | Keterangan |
|--------|-----|------|------------|
| `access_token` | 15 menit | `/` | JWT berisi user ID, role, sekolah_id |
| `refresh_token` | 7 hari | `/api/v1/auth/refresh` | Opaque token, hashed di DB |

Kedua cookie di-set dengan flag:
- `HttpOnly` — tidak bisa diakses JavaScript (anti-XSS)
- `Secure` — hanya dikirim via HTTPS
- `SameSite=Strict` — tidak dikirim cross-origin (anti-CSRF dasar)

### Login Flow

```
1. User kirim kode_sekolah + email + password
2. Resolve kode_sekolah → sekolah_id
   - Jika kode invalid → catat login_attempt(sekolah_id=NULL), return 401
3. Cek account lockout per (sekolah_id, email)
   - Jika locked → return 429 dengan info waktu unlock
4. Backend hash password dengan bcrypt, bandingkan dengan DB (scope: sekolah_id + email)
5. Jika gagal → catat di login_attempt(sekolah_id, email), cek threshold lockout
6. Jika cocok → generate access token (JWT, 15m) + refresh token (random, 7d)
7. Simpan hash refresh token di tabel refresh_token
8. Set dua httpOnly cookie (access_token + refresh_token)
9. Return user data
```

Lockout key:
- Sekolah valid: `(sekolah_id, email)` — isolasi per tenant
- Sekolah invalid: `(NULL, email)` + rate limit by IP — mencegah brute force kode sekolah

### Refresh Flow

```
1. Access token expired → frontend dapat 401
2. Frontend call POST /auth/refresh (cookie refresh_token otomatis terkirim)
3. Backend baca refresh token dari cookie
4. Cek hash token di DB: valid, belum expired, belum revoked
5. Generate access token baru + refresh token baru (rotation)
6. Revoke refresh token lama di DB
7. Set dua cookie baru
8. Return user data
```

Refresh token rotation memastikan token yang bocor hanya bisa dipakai sekali. Jika token lama dipakai setelah di-rotate, semua session user di-revoke (anomaly detection).

### Logout & Revoke

```
1. POST /auth/logout → revoke refresh token di DB, hapus kedua cookie
2. Admin bisa revoke semua session user via POST /auth/revoke-all/:user_id
```

### Password Policy

- Minimum 8 karakter
- Harus mengandung huruf dan angka
- Disimpan dengan bcrypt (salt otomatis)

### Token Refresh

Lihat "Refresh Flow" di atas. Catatan penting:

- Refresh token di-hash (SHA-256) sebelum disimpan di DB
- Setiap refresh menghasilkan token baru (rotation)
- Refresh token lama langsung di-revoke
- Jika token lama dipakai setelah rotation → revoke semua session user (breach detection)
- Path cookie refresh_token dibatasi ke `/api/v1/auth/refresh` agar tidak terkirim ke endpoint lain

### Account Lockout

Setelah beberapa kali gagal login, akun akan dikunci sementara:

| Percobaan Gagal | Aksi |
|-----------------|------|
| 3x | Delay 30 detik |
| 5x | Lockout 5 menit |
| 10x | Lockout 30 menit |
| 15x | Lockout permanen (perlu reset admin) |

Lockout di-track per email, bukan per IP, untuk mencegah brute force yang terdistribusi.

---

## Otorisasi

### RBAC (Role-Based Access Control)

| Role | Akses |
|------|-------|
| admin | Semua fitur, kelola user, pengaturan |
| operator | Pembayaran, PPDB, Siswa (tidak bisa hapus) |
| guru | Data kelas yang diajar, siswa di kelas |
| siswa | Data sendiri, tagihan, pembayaran |
| orangtua | Data anak, tagihan, pembayaran |

### Middleware

```go
// Contoh middleware otorisasi
func RequireRole(roles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := GetUserFromContext(r.Context())
            if !contains(roles, user.Role) {
                respond.Error(w, 403, "FORBIDDEN", "Tidak punya akses")
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Penggunaan
r.Group(func(r chi.Router) {
    r.Use(middleware.RequireRole("admin", "operator"))
    r.Get("/siswa", siswaHandler.List)
})
```

### Object-Level Authorization (Siswa & Orangtua)

Role `siswa` dan `orangtua` hanya boleh akses data siswa yang terhubung via tabel `pengguna_siswa`. Ini bukan role-level check saja — harus ada relasi eksplisit.

#### Flow Authorization

```
1. User dengan role siswa/orangtua request GET /siswa/:id/tagihan
2. Middleware cek role → lolos (role diizinkan)
3. Handler cek relasi:
   SELECT 1 FROM pengguna_siswa
   WHERE pengguna_id = ? AND siswa_id = ? AND aktif = TRUE
4. Jika tidak ada relasi → 404 (bukan 403, mencegah enumeration)
5. Jika ada → lanjut proses
```

#### Contoh Implementasi

```go
func (h *SiswaHandler) GetTagihan(w http.ResponseWriter, r *http.Request) {
    user := middleware.GetUser(r.Context())
    siswaID := chi.URLParam(r, "id")

    // Admin/Operator: cukup cek tenant (sekolah_id)
    if user.Role == "admin" || user.Role == "operator" {
        tagihan, err := h.service.GetTagihanBySiswa(r.Context(), user.SekolahID, siswaID)
        // ...
        return
    }

    // Siswa/Orangtua: cek relasi di pengguna_siswa
    hasAccess, err := h.service.CekRelasiPengguna(r.Context(), user.ID, siswaID)
    if !hasAccess {
        respond.Error(w, 404, "NOT_FOUND", "Data tidak ditemukan")
        return
    }

    tagihan, err := h.service.GetTagihanBySiswa(r.Context(), user.SekolahID, siswaID)
    // ...
}
```

#### Aturan Relasi

| Hubungan | Artinya |
|----------|---------|
| `diri_sendiri` | User role `siswa` terhubung ke record siswa miliknya |
| `ayah` | User role `orangtua` adalah ayah dari siswa |
| `ibu` | User role `orangtua` adalah ibu dari siswa |
| `wali` | User role `orangtua` adalah wali dari siswa |
| `lainnya` | Relasi lain yang diizinkan admin |

Satu orangtua bisa terhubung ke banyak siswa (misal kakak-adik). Satu siswa bisa punya banyak pengguna terhubung (ayah + ibu).

---

## CSRF Protection

Karena autentikasi menggunakan httpOnly cookie, aplikasi rentan terhadap CSRF. Proteksi yang diterapkan:

### Double Submit Cookie

```
1. Backend generate CSRF token saat login
2. Token dikirim sebagai non-httpOnly cookie (bisa dibaca JS)
3. Frontend baca token dari cookie
4. Frontend kirim token di header X-CSRF-Token setiap request mutasi (POST/PUT/DELETE)
5. Backend bandingkan header dengan cookie
6. Jika tidak cocok → 403 Forbidden
```

### Pengecualian CSRF

Endpoint yang tidak perlu CSRF check:
- `POST /api/v1/payment/callback/*` (validasi via signature dari payment gateway)
- `POST /api/v1/auth/login` (belum punya cookie)
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/refresh` (SameSite=Strict + path cookie terbatas, side effect hanya rotate token)

---

## Setup Guard

Endpoint setup (`GET /setup/status`, `POST /setup`) hanya aktif jika database belum initialized. Setelah setup selesai, endpoint return 404. Ini mencegah:
- Takeover: attacker re-run setup untuk buat admin baru
- Re-konfigurasi: ubah modul/notifikasi tanpa otorisasi

Mekanisme: middleware cek apakah tabel `sekolah` sudah punya data sebelum meneruskan request ke handler setup.

---

## Audit Log

Semua aksi sensitif dicatat di tabel `audit_log` untuk accountability:

### Aksi yang Dicatat

| Kategori | Aksi |
|----------|------|
| Auth | Login, logout, failed login, ganti password |
| Data | Hapus siswa, hapus tagihan, hapus kelas |
| Pembayaran | Verifikasi, tolak pembayaran |
| PPDB | Ubah status pendaftar, publish pengumuman |
| Sistem | Backup, restore, ubah konfigurasi |
| User | Buat user, ubah role, nonaktifkan user |

### Struktur Audit Log

```sql
CREATE TABLE audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    pengguna_id INTEGER,
    aksi TEXT NOT NULL,
    tabel TEXT,
    record_id INTEGER,
    data_lama TEXT,               -- JSON snapshot sebelum perubahan
    data_baru TEXT,               -- JSON snapshot setelah perubahan
    ip_address TEXT,
    user_agent TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    FOREIGN KEY (pengguna_id) REFERENCES pengguna(id)
);

CREATE INDEX idx_audit_log_pengguna_id ON audit_log(pengguna_id);
CREATE INDEX idx_audit_log_aksi ON audit_log(aksi);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);
```

### Retensi

- Default: simpan 90 hari
- Bisa dikonfigurasi di `config.yaml`
- Audit log lama dihapus otomatis oleh scheduler

---

## Proteksi

### Rate Limiting

| Endpoint | Limit | Window |
|----------|-------|--------|
| `POST /auth/login` | 5 request | per menit |
| `POST /auth/register` | 3 request | per menit |
| `POST /ppdb/daftar` | 10 request | per menit |
| Lainnya | 100 request | per menit |

### CORS

Hanya origin yang diizinkan di `config.yaml` yang bisa akses API.

```yaml
cors:
  allowed_origins:
    - "http://localhost:5173"    # Frontend dev
    - "http://localhost:8080"   # Production
```

### SQL Injection Prevention

Semua query menggunakan Squirrel dengan parameterized query:

```go
// ✅ Aman (parameterized)
sq.Select("*").From("siswa").Where(sq.Eq{"id": id})

// ❌ Berbahaya (raw SQL)
db.Query("SELECT * FROM siswa WHERE id = " + id)
```

### Input Validation

Semua input divalidasi di server:

```go
type CreateSiswaRequest struct {
    Nama   string `validate:"required,min=3,max=100"`
    NIS    string `validate:"required,alphanum"`
    Email  string `validate:"omitempty,email"`
    NoHP   string `validate:"omitempty,numeric"`
}
```

### File Upload Validation

- Tipe file: hanya JPEG, PNG, PDF
- Ukuran max: 5 MB
- Nama file: di-randomize untuk mencegah path traversal

### File Access Control

Semua file yang diupload dibagi dua kategori:

| Kategori | Contoh | Akses |
|----------|--------|-------|
| Public | Logo sekolah | Langsung via `/public/`, tanpa auth |
| Private | Bukti bayar, KK, akta, berkas PPDB, foto siswa | Hanya via auth handler `GET /api/v1/upload/:path` |

#### Default: Auth Handler

```
1. Request GET /api/v1/upload/:path
2. Cek auth (access_token cookie)
3. Cek role + relasi:
   - Admin/Operator: akses semua file di sekolah_id yang sama
   - Guru: akses foto siswa di kelas yang diajar
   - Siswa/Orangtua: akses file milik siswa yang terhubung (via pengguna_siswa)
4. Jika tidak punya akses → 404
5. Jika OK → stream file
```

#### Opsional: Signed URL

Untuk kasus di mana auth cookie tidak tersedia (link di notifikasi WhatsApp/Telegram/Email):

```
1. Backend generate signed URL: POST /api/v1/upload/signed
   - Request: { "path": "1/bukti_bayar/abc.pdf", "ttl_seconds": 300 }
   - Response: { "token": "...", "url": "/api/v1/upload/signed/...", "expires_at": "..." }
2. Token berisi: JSON payload (sekolah_id, path, expires_at) + HMAC-SHA256 signature
3. Format token: base64url(payload).base64url(hmac)
4. TTL: default 5 menit, max 15 menit (dipotong otomatis jika lebih)
5. Akses: GET /api/v1/upload/signed/:token — tidak perlu auth cookie
6. Validasi: HMAC signature, expiry timestamp, sekolah_id sesuai path, path traversal ditolak
7. Setelah expired → 403
8. Token reusable sampai expired (tidak single-use)
```

Signed URL digunakan untuk:
- Link download di pesan WhatsApp/Telegram
- Preview file di email notifikasi
- Sharing terbatas ke pihak ketiga

**Keamanan Signed URL:**
- Token ditandatangani dengan HMAC-SHA256 menggunakan JWT secret yang sama
- Payload berisi `sekolah_id` yang harus sesuai dengan path (mencegah cross-school access)
- Path traversal (`..`) dicegah di level validasi token
- Token yang diubah (tampered) akan gagal verifikasi signature
- Token yang sudah expired akan ditolak dengan 403
- Tidak ada data file yang bisa diakses tanpa token yang valid

#### Deployment

**Jangan** expose folder `/uploads/` langsung via Nginx/Caddy. Semua file privat harus melewati backend auth handler. Hanya folder `/public/` (logo, assets) yang boleh di-serve langsung.

---

## Data Protection

### Password

- Hashed dengan bcrypt (cost 10)
- Tidak pernah di-log atau di-return di API

### Sensitive Data

- JWT secret: minimal 32 karakter
- API keys: simpan di `.env`, bukan di `config.yaml`
- Database: file `.db` tidak boleh diakses publik

### Backup

- Backup dienkripsi (opsional, bisa ditambah nanti)
- Backup disimpan lokal, tidak dikirim ke server external

---

## Best Practices

### Deployment

1. Gunakan HTTPS (reverse proxy dengan Nginx/Caddy)
2. Jangan expose port database
3. Ganti JWT secret default
4. Matikan modul yang tidak digunakan

### Monitoring

1. Cek log secara berkala
2. Monitor failed login attempts
3. Monitor disk space (uploads & backups)

### Maintenance

1. Update dependencies secara berkala
2. Backup sebelum update
3. Test di staging sebelum production
