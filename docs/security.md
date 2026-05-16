# Keamanan

## Overview

Aplikasi menggunakan beberapa lapisan keamanan untuk melindungi data sekolah.

---

## Autentikasi

### JWT (JSON Web Token)

- Token disimpan di httpOnly cookie
- Masa berlaku: 24 jam
- Tidak bisa diakses via JavaScript (anti-XSS)

### Login Flow

```
1. User kirim email + password
2. Backend hash password dengan bcrypt
3. Bandingkan dengan hash di database
4. Jika cocok, generate JWT token
5. Set httpOnly cookie
6. Return user data
```

### Password Policy

- Minimum 8 karakter
- Harus mengandung huruf dan angka
- Disimpan dengan bcrypt (salt otomatis)

### Token Refresh

JWT access token memiliki masa berlaku pendek (15 menit). Untuk menjaga session tanpa login ulang:

```
1. Login → dapat access token (15 menit) + refresh token (7 hari)
2. Access token expired → frontend kirim refresh token
3. Backend validasi refresh token
4. Generate access token baru
5. Jika refresh token expired → user harus login ulang
```

Refresh token disimpan di database dan bisa di-revoke oleh admin.

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
