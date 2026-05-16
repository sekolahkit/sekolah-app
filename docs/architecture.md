# Arsitektur

## Overview

SekolahApp menggunakan arsitektur monorepo dengan backend Go dan frontend React yang di-embed ke dalam binary Go.

```
┌─────────────────────────────────────────┐
│              Binary Go                   │
│  ┌─────────────┐  ┌──────────────────┐  │
│  │  REST API    │  │  Frontend (SPA)  │  │
│  │  /api/v1/*   │  │  (embedded)      │  │
│  └──────┬───────┘  └──────────────────┘  │
│         │                                │
│  ┌──────┴───────┐  ┌──────────────────┐  │
│  │  Service      │  │  Database        │  │
│  │  Layer        │  │  (SQLite)        │  │
│  └──────┬───────┘  └──────────────────┘  │
│         │                                │
│  ┌──────┴───────┐  ┌──────────────────┐  │
│  │  Adapter      │  │  File Storage    │  │
│  │  (Payment,    │  │  (/uploads)      │  │
│  │   Notif)      │  │                  │  │
│  └──────────────┘  └──────────────────┘  │
└─────────────────────────────────────────┘
```

## Stack Teknologi

### Backend
| Komponen | Teknologi | Keterangan |
|----------|-----------|------------|
| Language | Go 1.22+ | |
| HTTP Router | Chi | Kompatibel net/http |
| Database | SQLite | File-based, ringan |
| Query Builder | Squirrel | Type-safe SQL builder |
| Migration | golang-migrate | Versioned migration |
| WhatsApp | whatsmeow | MPL-2.0, jangan modify |
| Logging | slog | Built-in Go 1.21+ |
| Excel | excelize | Import/Export Excel |

### Frontend
| Komponen | Teknologi | Keterangan |
|----------|-----------|------------|
| Framework | React 18+ | |
| Build Tool | Vite | |
| UI Library | shadcn/ui | Copy-paste components |
| State Server | TanStack Query | Fetching & caching |
| Table | TanStack Table | Headless table |
| Routing | React Router | SPA + lazy import |
| Styling | Tailwind CSS | |
| HTTP Client | axios | |

## Struktur Folder

### Backend (`/backend`)
```
/backend
  /cmd
    /server/main.go          # Entry point
  /internal
    /auth                    # Autentikasi & otorisasi
      handler.go
      service.go
      repository.go
    /siswa                   # Modul siswa
      handler.go
      service.go
      repository.go
    /kelas                   # Modul kelas
      handler.go
      service.go
      repository.go
    /pembayaran              # Modul pembayaran
      handler.go
      service.go
      repository.go
    /ppdb                    # Modul PPDB
      handler.go
      service.go
      repository.go
    /notifikasi              # Modul notifikasi
      handler.go
      service.go
      repository.go
      /adapter
        whatsapp.go
        telegram.go
        email.go
      /queue
        queue.go
    /backup                  # Modul backup
      handler.go
      service.go
    /setup                   # Setup wizard
      handler.go
      service.go
    /migration
      /migrations            # File SQL migration
  /pkg
    /database                # Koneksi database
    /middleware               # Middleware (auth, cors, rate limit)
    /response                # Response helper
    /validator               # Input validation
    /config                  # Config loader
```

### Frontend (`/frontend`)
```
/frontend
  /src
    /components
      /ui                    # shadcn/ui components
      /layout                # Sidebar, header, footer
    /features
      /auth
        LoginPage.tsx
        RegisterPage.tsx
      /dashboard
        DashboardPage.tsx
        AdminDashboard.tsx
        OperatorDashboard.tsx
        SiswaDashboard.tsx
      /siswa
        SiswaPage.tsx
        SiswaForm.tsx
        SiswaTable.tsx
      /kelas
        KelasPage.tsx
      /pembayaran
        PembayaranPage.tsx
        TagihanPage.tsx
        RiwayatPage.tsx
      /ppdb
        PendaftaranPage.tsx  # Publik
        VerifikasiPage.tsx   # Admin
        PengumumanPage.tsx   # Publik
      /laporan
        LaporanPembayaran.tsx
        LaporanPPDB.tsx
      /pengaturan
        PengaturanPage.tsx
      /setup
        SetupWizard.tsx
    /hooks                   # Custom hooks
    /lib                     # Utilities, API client
    /types                   # TypeScript types
  /public
```

## Keputusan Desain

### 1. Monorepo dengan Embed
Frontend di-embed ke binary Go menggunakan `//go:embed`. Hasil: satu file executable, deploy mudah.

### 2. Multi-tenancy & Tenant Scoping Contract

Setiap instalasi untuk satu sekolah (atau yayasan dengan beberapa sekolah). Selalu ada `sekolah_id` di tabel data.

#### Contract

Semua akses data **wajib** di-scope berdasarkan `sekolah_id` dari authenticated user. Ini bukan fitur opsional — ini invariant yang harus dipenuhi di setiap layer.

#### Mekanisme

```
1. Auth middleware extract user dari JWT → simpan di context
2. Context berisi: user_id, sekolah_id, role
3. Semua repository method menerima sekolah_id sebagai parameter wajib
4. Semua query di-filter WHERE sekolah_id = ?
5. Endpoint by :id → setelah fetch record, cek record.sekolah_id == user.sekolah_id
```

#### Repository Pattern

```go
// WAJIB: sekolah_id sebagai parameter pertama
func (r *SiswaRepository) FindByID(ctx context.Context, sekolahID int64, id int64) (*Siswa, error) {
    return sq.Select("*").
        From("siswa").
        Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
        RunWith(r.db).
        QueryRowContext(ctx)
}

// WAJIB: list selalu scope by sekolah_id
func (r *SiswaRepository) List(ctx context.Context, sekolahID int64, filter ListFilter) ([]Siswa, error) {
    return sq.Select("*").
        From("siswa").
        Where(sq.Eq{"sekolah_id": sekolahID}).
        RunWith(r.db).
        QueryContext(ctx)
}
```

#### Object-Level Authorization (Handler)

```go
func (h *SiswaHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    user := middleware.GetUser(r.Context())
    id := chi.URLParam(r, "id")

    siswa, err := h.service.FindByID(r.Context(), user.SekolahID, id)
    if err != nil {
        // FindByID sudah filter by sekolah_id
        // Jika tidak ditemukan → 404 (bukan 403, untuk mencegah enumeration)
        respond.Error(w, 404, "NOT_FOUND", "Siswa tidak ditemukan")
        return
    }

    respond.JSON(w, 200, siswa)
}
```

#### Aturan

| Layer | Tanggung Jawab |
|-------|----------------|
| Middleware | Extract `sekolah_id` dari JWT, simpan di context |
| Handler | Ambil `sekolah_id` dari context, pass ke service |
| Service | Pass `sekolah_id` ke repository, enforce business rules |
| Repository | **Selalu** filter query dengan `sekolah_id` |

> **Tidak boleh** ada query yang fetch record by ID tanpa filter `sekolah_id`. Ini mencegah data leak lintas sekolah meskipun attacker tahu ID record.

### 3. RBAC (Role-Based Access Control)
Role tetap: admin, operator, guru, siswa, orangtua. Setiap role punya izin berbeda.

### 4. Adapter Pattern
Untuk payment gateway dan notifikasi. Mudah tambah provider baru.

### 5. Modular PPDB
Sekolah bisa aktifkan/nonaktifkan modul PPDB sesuai kebutuhan.

### 6. Versioned Migration
Setiap perubahan schema punya file migration. Bisa upgrade dan rollback.

### 7. Auto Backup
Backup otomatis setiap hari, simpan 7 hari terakhir.

---

## Error Handling

### Strategi

Error di-handle secara konsisten di setiap layer dengan prinsip:
- **Repository layer**: Return error mentah dari database
- **Service layer**: Wrap error dengan konteks bisnis, buat custom error types
- **Handler layer**: Translate error ke HTTP response yang sesuai

### Custom Error Types

```go
type AppError struct {
    Code    string // VALIDATION_ERROR, NOT_FOUND, FORBIDDEN, dll
    Message string // Pesan untuk user
    Err     error  // Original error (tidak di-expose ke client)
}

type ValidationError struct {
    Field   string
    Message string
}
```

### Error Flow

```
Repository → error: "UNIQUE constraint failed: siswa.nis"
     ↓
Service → AppError{Code: "DUPLICATE", Message: "NIS sudah terdaftar"}
     ↓
Handler → HTTP 422 + JSON {"error": {"code": "DUPLICATE", "message": "NIS sudah terdaftar"}}
```

### Logging Strategy

| Layer | Yang Di-log |
|-------|-------------|
| Handler | Request method, path, status code, duration |
| Service | Business logic errors, state transitions |
| Repository | Query errors, slow queries (>100ms) |
| Adapter | External API calls, responses, timeouts |

Gunakan `slog` dengan structured logging:

```go
slog.Error("gagal verifikasi pembayaran",
    "pembayaran_id", id,
    "error", err,
    "user_id", userID,
)
```

### Panic Recovery

Middleware `recover` menangkap panic agar server tidak crash:

```go
func RecoverMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                slog.Error("panic recovered", "error", err, "stack", debug.Stack())
                respond.Error(w, 500, "INTERNAL_ERROR", "Terjadi kesalahan internal")
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```
