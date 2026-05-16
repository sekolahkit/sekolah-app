# Database Schema

## Overview

Database menggunakan SQLite. Setiap instalasi memiliki satu file database `sekolah.db`.

## Diagram Relasi

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   sekolahs   │────<│    users      │     │  tahun_     │
│              │     │              │     │  ajarans    │
└─────────────┘     └──────────────┘     └─────────────┘
       │                    │                    │
       │                    │                    │
       ▼                    ▼                    ▼
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   kelass     │────<│ kelas_siswas │>────│   siswas    │
│              │     │              │     │             │
└─────────────┘     └──────────────┘     └─────────────┘
                                                │
                                                │
       ┌────────────────────────────────────────┘
       │
       ▼
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  tagihans    │────<│ pembayarans  │     │ kategori_   │
│              │     │              │     │ pembayarans │
└─────────────┘     └──────────────┘     └─────────────┘
       │
       │
       ▼
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│ ppdb_        │────<│ ppdb_berkas  │     │ ppdb_ujians │
│ pendaftarans │     │              │     │             │
└─────────────┘     └──────────────┘     └─────────────┘
       │
       │
       ▼
┌─────────────┐     ┌──────────────┐
│ ppdb_        │     │ notifikasi_  │
│ pengumuman   │     │ queue        │
└─────────────┘     └──────────────┘
```

## Tabel-tabel

### sekolahs
Tabel utama untuk data sekolah.

```sql
CREATE TABLE sekolahs (
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

### users
Tabel pengguna sistem dengan role-based access.

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,        -- bcrypt hash
    nama TEXT NOT NULL,
    role TEXT NOT NULL,            -- admin, operator, guru, siswa, orangtua
    google_id TEXT,                -- untuk login Google
    foto TEXT,
    no_hp TEXT,
    aktif BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolahs(id)
);
```

### tahun_ajarans
Tabel tahun ajaran. Terpisah dari semester.

```sql
CREATE TABLE tahun_ajarans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    nama TEXT NOT NULL,            -- "2024/2025"
    aktif BOOLEAN DEFAULT FALSE,
    tanggal_mulai DATE,
    tanggal_selesai DATE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolahs(id)
);
```

### kelass
Tabel kelas dengan tingkat SD 1-6, SMP 7-9, SMA/SMK 10-12.

```sql
CREATE TABLE kelass (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    nama TEXT NOT NULL,            -- "7A", "X TKJ 1"
    tingkat INTEGER NOT NULL,      -- 1-12
    jurusan_id INTEGER,            -- opsional, untuk SMK
    wali_kelas_id INTEGER,         -- referensi ke users
    ruangan TEXT,
    kapasitas INTEGER,
    shift TEXT,                    -- "pagi", "sore"
    tahun_ajaran_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolahs(id),
    FOREIGN KEY (tahun_ajaran_id) REFERENCES tahun_ajarans(id),
    FOREIGN KEY (wali_kelas_id) REFERENCES users(id)
);
```

### jurusans
Tabel jurusan untuk SMK.

```sql
CREATE TABLE jurusans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    nama TEXT NOT NULL,            -- "TKJ", "RPL", "Akuntansi"
    kode TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolahs(id)
);
```

### siswas
Tabel data siswa. Tidak punya kelas_id langsung (gunakan tabel penghubung).

```sql
CREATE TABLE siswas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    nis TEXT NOT NULL UNIQUE,
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
    FOREIGN KEY (sekolah_id) REFERENCES sekolahs(id)
);
```

### kelas_siswas
Tabel penghubung siswa dan kelas per tahun ajaran. Mendukung tracking mutasi.

```sql
CREATE TABLE kelas_siswas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    siswa_id INTEGER NOT NULL,
    kelas_id INTEGER NOT NULL,
    tahun_ajaran_id INTEGER NOT NULL,
    status TEXT DEFAULT 'aktif',   -- aktif, lulus, pindah, keluar
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (siswa_id) REFERENCES siswas(id),
    FOREIGN KEY (kelas_id) REFERENCES kelass(id),
    FOREIGN KEY (tahun_ajaran_id) REFERENCES tahun_ajarans(id)
);
```

### kategori_pembayarans
Kategori pembayaran yang bisa dikustomisasi per sekolah.

```sql
CREATE TABLE kategori_pembayarans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    nama TEXT NOT NULL,            -- "SPP", "Uang Pangkal", "Seragam"
    deskripsi TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolahs(id)
);
```

### tagihans
Tabel tagihan per siswa. Support cicilan (banyak pembayaran per tagihan).

```sql
CREATE TABLE tagihans (
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
    FOREIGN KEY (sekolah_id) REFERENCES sekolahs(id),
    FOREIGN KEY (siswa_id) REFERENCES siswas(id),
    FOREIGN KEY (kategori_id) REFERENCES kategori_pembayarans(id),
    FOREIGN KEY (tahun_ajaran_id) REFERENCES tahun_ajarans(id)
);
```

### pembayarans
Tabel pembayaran/cicilan per tagihan.

```sql
CREATE TABLE pembayarans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tagihan_id INTEGER NOT NULL,
    siswa_id INTEGER NOT NULL,
    jumlah DECIMAL(15,2) NOT NULL,
    tanggal DATE NOT NULL,
    metode TEXT NOT NULL,          -- transfer, cash, midtrans, xendit
    bukti_bayar TEXT,              -- path file
    payment_gateway_id TEXT,       -- ID dari payment gateway
    status TEXT DEFAULT 'pending', -- pending, verified, rejected
    catatan TEXT,
    verified_by INTEGER,           -- user ID yang verifikasi
    verified_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tagihan_id) REFERENCES tagihans(id),
    FOREIGN KEY (siswa_id) REFERENCES siswas(id),
    FOREIGN KEY (verified_by) REFERENCES users(id)
);
```

### ppdb_pendaftarans
Tabel pendaftar PPDB.

```sql
CREATE TABLE ppdb_pendaftarans (
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
    FOREIGN KEY (sekolah_id) REFERENCES sekolahs(id),
    FOREIGN KEY (tahun_ajaran_id) REFERENCES tahun_ajarans(id)
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
    FOREIGN KEY (pendaftaran_id) REFERENCES ppdb_pendaftarans(id)
);
```

### ppdb_ujians
Tabel nilai ujian PPDB (modul opsional).

```sql
CREATE TABLE ppdb_ujians (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pendaftaran_id INTEGER NOT NULL,
    nama_ujian TEXT NOT NULL,
    nilai DECIMAL(5,2),
    keterangan TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pendaftaran_id) REFERENCES ppdb_pendaftarans(id)
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
    FOREIGN KEY (pendaftaran_id) REFERENCES ppdb_pendaftarans(id)
);
```

### notifikasi_queue
Tabel antrian notifikasi untuk retry.

```sql
CREATE TABLE notifikasi_queue (
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
    FOREIGN KEY (sekolah_id) REFERENCES sekolahs(id)
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
