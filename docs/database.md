# Database Schema

## Overview

Database menggunakan SQLite. Setiap instalasi memiliki satu file database `sekolah.db`.

## Diagram Relasi

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   sekolah    │────<│   pengguna   │     │  tahun_     │
│              │     │              │     │  ajaran     │
└─────────────┘     └──────────────┘     └─────────────┘
       │                    │                    │
       │                    │                    │
       ▼                    ▼                    ▼
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   kelas      │────<│ kelas_siswa  │>────│   siswa     │
│              │     │              │     │             │
└─────────────┘     └──────────────┘     └─────────────┘
                                                │
                                                │
       ┌────────────────────────────────────────┘
       │
       ▼
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  tagihan     │────<│  pembayaran  │     │ kategori_   │
│              │     │              │     │ pembayaran  │
└─────────────┘     └──────────────┘     └─────────────┘
       │
       │
       ▼
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│ ppdb_        │────<│ ppdb_berkas  │     │ ppdb_ujian  │
│ pendaftaran  │     │              │     │             │
└─────────────┘     └──────────────┘     └─────────────┘
       │
       │
       ▼
┌─────────────┐     ┌──────────────┐
│ ppdb_        │     │ notifikasi_  │
│ pengumuman   │     │ antrian      │
└─────────────┘     └──────────────┘
```

## Konvensi Penamaan

- Nama tabel: bahasa Indonesia, singular, snake_case
- Nama kolom: bahasa Indonesia, snake_case
- Foreign key: `<tabel_referensi>_id`
- Tabel penghubung: `<tabel1>_<tabel2>` (contoh: `kelas_siswa`)

## Tabel-tabel

### sekolah
Tabel utama untuk data sekolah.

```sql
CREATE TABLE sekolah (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nama TEXT NOT NULL,
    alamat TEXT,
    telepon TEXT,
    email TEXT,
    logo TEXT,                    -- path file logo
    website TEXT,
    nama_kepala_sekolah TEXT,
    nip_kepala_sekolah TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### pengguna
Tabel pengguna sistem dengan role-based access.

```sql
CREATE TABLE pengguna (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL,        -- bcrypt hash
    nama TEXT NOT NULL,
    role TEXT NOT NULL,            -- admin, operator, guru, siswa, orangtua
    google_id TEXT,                -- untuk login Google
    foto TEXT,
    no_hp TEXT,
    aktif BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    UNIQUE(sekolah_id, email)
);
```

### refresh_token
Tabel refresh token untuk session management. Token di-hash sebelum disimpan.

```sql
CREATE TABLE refresh_token (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pengguna_id INTEGER NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,  -- SHA-256 hash
    device_info TEXT,                  -- user agent / device identifier
    expires_at DATETIME NOT NULL,
    revoked_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pengguna_id) REFERENCES pengguna(id)
);
```

### login_attempt
Tabel tracking percobaan login untuk account lockout.

```sql
CREATE TABLE login_attempt (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL,
    ip_address TEXT,
    success BOOLEAN NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### tahun_ajaran
Tabel tahun ajaran. Terpisah dari semester.

```sql
CREATE TABLE tahun_ajaran (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    nama TEXT NOT NULL,            -- "2024/2025"
    aktif BOOLEAN DEFAULT FALSE,
    tanggal_mulai DATE,
    tanggal_selesai DATE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id)
);
```

### kelas
Tabel kelas dengan tingkat SD 1-6, SMP 7-9, SMA/SMK 10-12.

```sql
CREATE TABLE kelas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    nama TEXT NOT NULL,            -- "7A", "X TKJ 1"
    tingkat INTEGER NOT NULL,      -- 1-12
    jurusan_id INTEGER,            -- opsional, untuk SMK
    wali_kelas_id INTEGER,         -- referensi ke pengguna
    ruangan TEXT,
    kapasitas INTEGER,
    shift TEXT,                    -- "pagi", "sore"
    tahun_ajaran_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    FOREIGN KEY (tahun_ajaran_id) REFERENCES tahun_ajaran(id),
    FOREIGN KEY (wali_kelas_id) REFERENCES pengguna(id)
);
```

### jurusan
Tabel jurusan untuk SMK.

```sql
CREATE TABLE jurusan (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    nama TEXT NOT NULL,            -- "TKJ", "RPL", "Akuntansi"
    kode TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id)
);
```

### siswa
Tabel data siswa. Tidak punya kelas_id langsung (gunakan tabel penghubung).

```sql
CREATE TABLE siswa (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    nis TEXT NOT NULL,
    nama TEXT NOT NULL,
    jenis_kelamin TEXT NOT NULL,   -- "L" / "P"
    tanggal_lahir DATE,
    tempat_lahir TEXT,
    agama TEXT,
    alamat TEXT,
    no_hp TEXT,
    email TEXT,
    foto TEXT,
    nama_ortu TEXT,
    no_hp_ortu TEXT,
    email_ortu TEXT,
    tahun_ajaran_masuk INTEGER,
    status TEXT DEFAULT 'aktif',   -- aktif, lulus, pindah, keluar
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    UNIQUE(sekolah_id, nis)
);
```

### pengguna_siswa
Tabel relasi pengguna ke siswa untuk authorization. Menentukan siapa boleh akses data siswa mana.

```sql
CREATE TABLE pengguna_siswa (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pengguna_id INTEGER NOT NULL,
    siswa_id INTEGER NOT NULL,
    hubungan TEXT NOT NULL,        -- diri_sendiri, ayah, ibu, wali, lainnya
    aktif BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pengguna_id) REFERENCES pengguna(id),
    FOREIGN KEY (siswa_id) REFERENCES siswa(id),
    UNIQUE(pengguna_id, siswa_id)
);
```

> **Catatan:** Field `nama_ortu`, `no_hp_ortu`, `email_ortu` di tabel `siswa` tetap dipertahankan sebagai data kontak denormalized. Tabel `pengguna_siswa` adalah sumber kebenaran untuk authorization, bukan field kontak tersebut.

### kelas_siswa
Tabel penghubung siswa dan kelas per tahun ajaran. Mendukung tracking mutasi.

```sql
CREATE TABLE kelas_siswa (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    siswa_id INTEGER NOT NULL,
    kelas_id INTEGER NOT NULL,
    tahun_ajaran_id INTEGER NOT NULL,
    status TEXT DEFAULT 'aktif',   -- aktif, lulus, pindah, keluar
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    FOREIGN KEY (siswa_id) REFERENCES siswa(id),
    FOREIGN KEY (kelas_id) REFERENCES kelas(id),
    FOREIGN KEY (tahun_ajaran_id) REFERENCES tahun_ajaran(id)
);
```

### kategori_pembayaran
Kategori pembayaran yang bisa dikustomisasi per sekolah.

```sql
CREATE TABLE kategori_pembayaran (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    nama TEXT NOT NULL,            -- "SPP", "Uang Pangkal", "Seragam"
    deskripsi TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id)
);
```

### tagihan
Tabel tagihan per siswa. Support cicilan (banyak pembayaran per tagihan).

```sql
CREATE TABLE tagihan (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    siswa_id INTEGER NOT NULL,
    kategori_id INTEGER NOT NULL,
    tahun_ajaran_id INTEGER NOT NULL,
    semester TEXT,                 -- opsional: "Ganjil", "Genap"
    nominal DECIMAL(15,2) NOT NULL,
    jatuh_tempo DATE,
    status TEXT DEFAULT 'belum_bayar',  -- belum_bayar, sebagian, lunas
    catatan TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    FOREIGN KEY (siswa_id) REFERENCES siswa(id),
    FOREIGN KEY (kategori_id) REFERENCES kategori_pembayaran(id),
    FOREIGN KEY (tahun_ajaran_id) REFERENCES tahun_ajaran(id)
);
```

### pembayaran
Tabel pembayaran/cicilan per tagihan.

```sql
CREATE TABLE pembayaran (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tagihan_id INTEGER NOT NULL,
    siswa_id INTEGER NOT NULL,
    jumlah DECIMAL(15,2) NOT NULL,
    tanggal DATE NOT NULL,
    metode TEXT NOT NULL,          -- transfer, cash, midtrans, xendit
    provider TEXT,                  -- midtrans, xendit (NULL jika manual)
    bukti_bayar TEXT,              -- path file
    payment_gateway_id TEXT,       -- ID dari payment gateway
    idempotency_key TEXT,          -- unique key untuk mencegah double processing
    status TEXT DEFAULT 'pending', -- pending, verified, rejected
    catatan TEXT,
    verified_by INTEGER,           -- pengguna ID yang verifikasi
    verified_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tagihan_id) REFERENCES tagihan(id),
    FOREIGN KEY (siswa_id) REFERENCES siswa(id),
    FOREIGN KEY (verified_by) REFERENCES pengguna(id),
    UNIQUE(provider, payment_gateway_id)
);
```

> **Invariant:** `SUM(pembayaran.jumlah WHERE status='verified' AND tagihan_id=X) <= tagihan.nominal`. Pembayaran yang menyebabkan overpay harus ditolak di level service dalam DB transaction.

### ppdb_pendaftaran
Tabel pendaftar PPDB.

```sql
CREATE TABLE ppdb_pendaftaran (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    tahun_ajaran_id INTEGER NOT NULL,
    nama_lengkap TEXT NOT NULL,
    nik TEXT,
    tempat_lahir TEXT,
    tanggal_lahir DATE,
    jenis_kelamin TEXT NOT NULL,
    agama TEXT,
    alamat TEXT,
    asal_sekolah TEXT,
    no_hp TEXT,
    email TEXT,
    nama_ortu TEXT,
    no_hp_ortu TEXT,
    pekerjaan_ortu TEXT,
    foto TEXT,
    status TEXT DEFAULT 'menunggu',
    -- menunggu, berkas_lengkap, berkas_ditolak,
    -- diterima, tidak_diterima, daftar_ulang
    catatan TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    FOREIGN KEY (tahun_ajaran_id) REFERENCES tahun_ajaran(id)
);
```

### ppdb_berkas
Tabel berkas yang diupload pendaftar PPDB.

```sql
CREATE TABLE ppdb_berkas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pendaftaran_id INTEGER NOT NULL,
    jenis_berkas TEXT NOT NULL,
    -- ijazah, skhun, akta_lahir, kk, foto, lainnya
    file_path TEXT NOT NULL,
    status TEXT DEFAULT 'pending', -- pending, diterima, ditolak
    catatan TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pendaftaran_id) REFERENCES ppdb_pendaftaran(id)
);
```

### ppdb_ujian
Tabel nilai ujian PPDB (modul opsional).

```sql
CREATE TABLE ppdb_ujian (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pendaftaran_id INTEGER NOT NULL,
    nama_ujian TEXT NOT NULL,
    nilai DECIMAL(5,2),
    keterangan TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pendaftaran_id) REFERENCES ppdb_pendaftaran(id)
);
```

### ppdb_pengumuman
Tabel pengumuman PPDB.

```sql
CREATE TABLE ppdb_pengumuman (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pendaftaran_id INTEGER NOT NULL,
    status TEXT NOT NULL,          -- diterima, tidak_diterima
    ranking INTEGER,
    keterangan TEXT,
    tanggal_pengumuman DATE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pendaftaran_id) REFERENCES ppdb_pendaftaran(id)
);
```

### notifikasi_antrian
Tabel antrian notifikasi untuk retry.

```sql
CREATE TABLE notifikasi_antrian (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    tipe TEXT NOT NULL,            -- whatsapp, telegram, email
    penerima TEXT NOT NULL,
    pesan TEXT NOT NULL,
    status TEXT DEFAULT 'pending', -- pending, sent, failed
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    last_error TEXT,
    scheduled_at DATETIME,
    sent_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id)
);
```

### schema_migrations
Tabel tracking versi migration.

```sql
CREATE TABLE schema_migrations (
    version INTEGER NOT NULL,
    dirty BOOLEAN NOT NULL
);
```

## Status Enums

### Tagihan Status
| Status | Keterangan |
|--------|------------|
| `belum_bayar` | Belum ada pembayaran |
| `sebagian` | Sudah bayar sebagian (cicilan) |
| `lunas` | Sudah lunas |

### Pembayaran Status
| Status | Keterangan |
|--------|------------|
| `pending` | Menunggu verifikasi |
| `verified` | Sudah diverifikasi |
| `rejected` | Ditolak |

### PPDB Status
| Status | Keterangan |
|--------|------------|
| `menunggu` | Baru daftar |
| `berkas_lengkap` | Berkas sudah lengkap |
| `berkas_ditolak` | Berkas ditolak |
| `diterima` | Diterima |
| `tidak_diterima` | Tidak diterima |
| `daftar_ulang` | Sudah daftar ulang |

### Siswa Status
| Status | Keterangan |
|--------|------------|
| `aktif` | Masih aktif |
| `lulus` | Sudah lulus |
| `pindah` | Pindah sekolah |
| `keluar` | Keluar/MOS |

### User Roles
| Role | Keterangan |
|------|------------|
| `admin` | Akses penuh |
| `operator` | Operator keuangan/PPDB |
| `guru` | Akses terbatas |
| `siswa` | Akses data sendiri |
| `orangtua` | Akses data anak |

---

## Indexes

Index yang direkomendasikan untuk performa query optimal:

```sql
-- pengguna
CREATE INDEX idx_pengguna_sekolah_id ON pengguna(sekolah_id);
CREATE INDEX idx_pengguna_email ON pengguna(email);
CREATE INDEX idx_pengguna_role ON pengguna(role);
CREATE INDEX idx_pengguna_google_id ON pengguna(google_id);

-- siswa
CREATE INDEX idx_siswa_sekolah_id ON siswa(sekolah_id);
CREATE INDEX idx_siswa_nis ON siswa(nis);
CREATE INDEX idx_siswa_nama ON siswa(nama);
CREATE INDEX idx_siswa_status ON siswa(status);

-- kelas
CREATE INDEX idx_kelas_sekolah_id ON kelas(sekolah_id);
CREATE INDEX idx_kelas_tahun_ajaran_id ON kelas(tahun_ajaran_id);
CREATE INDEX idx_kelas_wali_kelas_id ON kelas(wali_kelas_id);

-- kelas_siswa
CREATE INDEX idx_kelas_siswa_siswa_id ON kelas_siswa(siswa_id);
CREATE INDEX idx_kelas_siswa_kelas_id ON kelas_siswa(kelas_id);
CREATE INDEX idx_kelas_siswa_tahun_ajaran_id ON kelas_siswa(tahun_ajaran_id);
CREATE UNIQUE INDEX idx_kelas_siswa_unique ON kelas_siswa(siswa_id, kelas_id, tahun_ajaran_id);

-- tagihan
CREATE INDEX idx_tagihan_siswa_id ON tagihan(siswa_id);
CREATE INDEX idx_tagihan_status ON tagihan(status);
CREATE INDEX idx_tagihan_jatuh_tempo ON tagihan(jatuh_tempo);
CREATE INDEX idx_tagihan_tahun_ajaran_id ON tagihan(tahun_ajaran_id);
CREATE INDEX idx_tagihan_kategori_id ON tagihan(kategori_id);

-- pembayaran
CREATE INDEX idx_pembayaran_tagihan_id ON pembayaran(tagihan_id);
CREATE INDEX idx_pembayaran_siswa_id ON pembayaran(siswa_id);
CREATE INDEX idx_pembayaran_status ON pembayaran(status);
CREATE INDEX idx_pembayaran_tanggal ON pembayaran(tanggal);
CREATE UNIQUE INDEX idx_pembayaran_gateway ON pembayaran(provider, payment_gateway_id);

-- refresh_token
CREATE INDEX idx_refresh_token_pengguna_id ON refresh_token(pengguna_id);
CREATE INDEX idx_refresh_token_expires_at ON refresh_token(expires_at);

-- login_attempt
CREATE INDEX idx_login_attempt_email ON login_attempt(email);
CREATE INDEX idx_login_attempt_created_at ON login_attempt(created_at);

-- ppdb_pendaftaran
CREATE INDEX idx_ppdb_pendaftaran_sekolah_id ON ppdb_pendaftaran(sekolah_id);
CREATE INDEX idx_ppdb_pendaftaran_tahun_ajaran_id ON ppdb_pendaftaran(tahun_ajaran_id);
CREATE INDEX idx_ppdb_pendaftaran_status ON ppdb_pendaftaran(status);

-- ppdb_berkas
CREATE INDEX idx_ppdb_berkas_pendaftaran_id ON ppdb_berkas(pendaftaran_id);

-- notifikasi_antrian
CREATE INDEX idx_notifikasi_antrian_status ON notifikasi_antrian(status);
CREATE INDEX idx_notifikasi_antrian_scheduled_at ON notifikasi_antrian(scheduled_at);

-- kelas_siswa (tenant)
CREATE INDEX idx_kelas_siswa_sekolah_id ON kelas_siswa(sekolah_id);
```
