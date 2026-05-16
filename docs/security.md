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
