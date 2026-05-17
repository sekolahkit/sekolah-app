CREATE TABLE sekolah (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nama TEXT NOT NULL,
    kode TEXT NOT NULL UNIQUE,
    alamat TEXT,
    telepon TEXT,
    email TEXT,
    logo TEXT,
    website TEXT,
    nama_kepala_sekolah TEXT,
    nip_kepala_sekolah TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE pengguna (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    nama TEXT NOT NULL,
    role TEXT NOT NULL,
    google_id TEXT,
    foto TEXT,
    no_hp TEXT,
    aktif BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    UNIQUE(sekolah_id, email)
);

CREATE TABLE refresh_token (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pengguna_id INTEGER NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    device_info TEXT,
    expires_at DATETIME NOT NULL,
    revoked_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pengguna_id) REFERENCES pengguna(id)
);

CREATE TABLE login_attempt (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER,
    email TEXT NOT NULL,
    ip_address TEXT,
    success BOOLEAN NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id)
);

CREATE TABLE tahun_ajaran (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    nama TEXT NOT NULL,
    aktif BOOLEAN DEFAULT FALSE,
    tanggal_mulai DATE,
    tanggal_selesai DATE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id)
);

CREATE TABLE jurusan (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    nama TEXT NOT NULL,
    kode TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id)
);

CREATE TABLE kelas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    nama TEXT NOT NULL,
    tingkat INTEGER NOT NULL,
    jurusan_id INTEGER,
    wali_kelas_id INTEGER,
    ruangan TEXT,
    kapasitas INTEGER,
    shift TEXT,
    tahun_ajaran_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    FOREIGN KEY (tahun_ajaran_id) REFERENCES tahun_ajaran(id),
    FOREIGN KEY (wali_kelas_id) REFERENCES pengguna(id)
);

CREATE TABLE siswa (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    nis TEXT NOT NULL,
    nama TEXT NOT NULL,
    jenis_kelamin TEXT NOT NULL,
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
    status TEXT DEFAULT 'aktif',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    UNIQUE(sekolah_id, nis)
);

CREATE TABLE pengguna_siswa (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    pengguna_id INTEGER NOT NULL,
    siswa_id INTEGER NOT NULL,
    hubungan TEXT NOT NULL,
    aktif BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    FOREIGN KEY (pengguna_id) REFERENCES pengguna(id),
    FOREIGN KEY (siswa_id) REFERENCES siswa(id),
    UNIQUE(sekolah_id, pengguna_id, siswa_id)
);

CREATE TABLE kelas_siswa (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    siswa_id INTEGER NOT NULL,
    kelas_id INTEGER NOT NULL,
    tahun_ajaran_id INTEGER NOT NULL,
    status TEXT DEFAULT 'aktif',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    FOREIGN KEY (siswa_id) REFERENCES siswa(id),
    FOREIGN KEY (kelas_id) REFERENCES kelas(id),
    FOREIGN KEY (tahun_ajaran_id) REFERENCES tahun_ajaran(id)
);

CREATE TABLE kategori_pembayaran (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    nama TEXT NOT NULL,
    deskripsi TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id)
);

CREATE TABLE rekening_sekolah (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    nama_bank TEXT NOT NULL,
    nomor_rekening TEXT NOT NULL,
    nama_pemilik TEXT NOT NULL,
    cabang TEXT,
    aktif BOOLEAN DEFAULT TRUE,
    urutan INTEGER DEFAULT 0,
    catatan TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    UNIQUE(sekolah_id, nama_bank, nomor_rekening)
);

CREATE TABLE tagihan (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    siswa_id INTEGER NOT NULL,
    kategori_id INTEGER NOT NULL,
    tahun_ajaran_id INTEGER NOT NULL,
    semester TEXT,
    nominal DECIMAL(15,2) NOT NULL,
    jatuh_tempo DATE,
    status TEXT DEFAULT 'belum_bayar',
    catatan TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    FOREIGN KEY (siswa_id) REFERENCES siswa(id),
    FOREIGN KEY (kategori_id) REFERENCES kategori_pembayaran(id),
    FOREIGN KEY (tahun_ajaran_id) REFERENCES tahun_ajaran(id)
);

CREATE TABLE pembayaran (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tagihan_id INTEGER NOT NULL,
    siswa_id INTEGER NOT NULL,
    jumlah DECIMAL(15,2) NOT NULL,
    tanggal DATE NOT NULL,
    metode TEXT NOT NULL,
    provider TEXT,
    bukti_bayar TEXT,
    payment_gateway_id TEXT,
    rekening_sekolah_id INTEGER,
    status TEXT DEFAULT 'pending',
    catatan TEXT,
    verified_by INTEGER,
    verified_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tagihan_id) REFERENCES tagihan(id),
    FOREIGN KEY (siswa_id) REFERENCES siswa(id),
    FOREIGN KEY (verified_by) REFERENCES pengguna(id),
    FOREIGN KEY (rekening_sekolah_id) REFERENCES rekening_sekolah(id),
    UNIQUE(provider, payment_gateway_id)
);

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
    skor DECIMAL(10,2),
    ranking INTEGER,
    latitude REAL,
    longitude REAL,
    catatan TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    FOREIGN KEY (tahun_ajaran_id) REFERENCES tahun_ajaran(id)
);

CREATE TABLE ppdb_berkas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pendaftaran_id INTEGER NOT NULL,
    jenis_berkas TEXT NOT NULL,
    file_path TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    catatan TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pendaftaran_id) REFERENCES ppdb_pendaftaran(id)
);

CREATE TABLE ppdb_ujian (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pendaftaran_id INTEGER NOT NULL,
    nama_ujian TEXT NOT NULL,
    nilai DECIMAL(5,2),
    keterangan TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pendaftaran_id) REFERENCES ppdb_pendaftaran(id)
);

CREATE TABLE ppdb_pengumuman (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pendaftaran_id INTEGER NOT NULL,
    status TEXT NOT NULL,
    ranking INTEGER,
    keterangan TEXT,
    tanggal_pengumuman DATE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pendaftaran_id) REFERENCES ppdb_pendaftaran(id)
);

CREATE TABLE ppdb_konfigurasi_ranking (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    tahun_ajaran_id INTEGER NOT NULL,
    metode TEXT NOT NULL,
    bobot_json TEXT,
    kuota INTEGER NOT NULL,
    cadangan INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    FOREIGN KEY (tahun_ajaran_id) REFERENCES tahun_ajaran(id),
    UNIQUE(sekolah_id, tahun_ajaran_id)
);

CREATE TABLE notifikasi_antrian (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    tipe TEXT NOT NULL,
    penerima TEXT NOT NULL,
    pesan TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    last_error TEXT,
    scheduled_at DATETIME,
    sent_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id)
);

CREATE TABLE audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    pengguna_id INTEGER,
    aksi TEXT NOT NULL,
    tabel TEXT,
    record_id INTEGER,
    data_lama TEXT,
    data_baru TEXT,
    ip_address TEXT,
    user_agent TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    FOREIGN KEY (pengguna_id) REFERENCES pengguna(id)
);

-- Indexes
CREATE INDEX idx_pengguna_sekolah_id ON pengguna(sekolah_id);
CREATE INDEX idx_pengguna_role ON pengguna(role);
CREATE INDEX idx_pengguna_google_id ON pengguna(google_id);

CREATE INDEX idx_refresh_token_pengguna_id ON refresh_token(pengguna_id);
CREATE INDEX idx_refresh_token_expires_at ON refresh_token(expires_at);

CREATE INDEX idx_login_attempt_sekolah_email ON login_attempt(sekolah_id, email);
CREATE INDEX idx_login_attempt_created_at ON login_attempt(created_at);
CREATE INDEX idx_login_attempt_ip ON login_attempt(ip_address);

CREATE INDEX idx_siswa_sekolah_id ON siswa(sekolah_id);
CREATE INDEX idx_siswa_nama ON siswa(nama);
CREATE INDEX idx_siswa_status ON siswa(status);

CREATE INDEX idx_kelas_sekolah_id ON kelas(sekolah_id);
CREATE INDEX idx_kelas_tahun_ajaran_id ON kelas(tahun_ajaran_id);
CREATE INDEX idx_kelas_wali_kelas_id ON kelas(wali_kelas_id);

CREATE INDEX idx_kelas_siswa_sekolah_id ON kelas_siswa(sekolah_id);
CREATE INDEX idx_kelas_siswa_siswa_id ON kelas_siswa(siswa_id);
CREATE INDEX idx_kelas_siswa_kelas_id ON kelas_siswa(kelas_id);
CREATE INDEX idx_kelas_siswa_tahun_ajaran_id ON kelas_siswa(tahun_ajaran_id);
CREATE UNIQUE INDEX idx_kelas_siswa_unique ON kelas_siswa(sekolah_id, siswa_id, kelas_id, tahun_ajaran_id);

CREATE INDEX idx_pengguna_siswa_sekolah_id ON pengguna_siswa(sekolah_id);
CREATE INDEX idx_pengguna_siswa_pengguna_id ON pengguna_siswa(pengguna_id);
CREATE INDEX idx_pengguna_siswa_siswa_id ON pengguna_siswa(siswa_id);

CREATE INDEX idx_rekening_sekolah_sekolah_id ON rekening_sekolah(sekolah_id);
CREATE INDEX idx_rekening_sekolah_aktif ON rekening_sekolah(aktif);

CREATE INDEX idx_tagihan_siswa_id ON tagihan(siswa_id);
CREATE INDEX idx_tagihan_status ON tagihan(status);
CREATE INDEX idx_tagihan_jatuh_tempo ON tagihan(jatuh_tempo);
CREATE INDEX idx_tagihan_tahun_ajaran_id ON tagihan(tahun_ajaran_id);
CREATE INDEX idx_tagihan_kategori_id ON tagihan(kategori_id);

CREATE INDEX idx_pembayaran_tagihan_id ON pembayaran(tagihan_id);
CREATE INDEX idx_pembayaran_siswa_id ON pembayaran(siswa_id);
CREATE INDEX idx_pembayaran_status ON pembayaran(status);
CREATE INDEX idx_pembayaran_tanggal ON pembayaran(tanggal);
CREATE UNIQUE INDEX idx_pembayaran_gateway ON pembayaran(provider, payment_gateway_id);
CREATE INDEX idx_pembayaran_rekening_id ON pembayaran(rekening_sekolah_id);

CREATE INDEX idx_ppdb_pendaftaran_sekolah_id ON ppdb_pendaftaran(sekolah_id);
CREATE INDEX idx_ppdb_pendaftaran_tahun_ajaran_id ON ppdb_pendaftaran(tahun_ajaran_id);
CREATE INDEX idx_ppdb_pendaftaran_status ON ppdb_pendaftaran(status);
CREATE INDEX idx_ppdb_pendaftaran_ranking ON ppdb_pendaftaran(ranking);
CREATE INDEX idx_ppdb_pendaftaran_skor ON ppdb_pendaftaran(skor);

CREATE INDEX idx_ppdb_berkas_pendaftaran_id ON ppdb_berkas(pendaftaran_id);

CREATE INDEX idx_ppdb_konfigurasi_ranking_sekolah_id ON ppdb_konfigurasi_ranking(sekolah_id);

CREATE INDEX idx_notifikasi_antrian_status ON notifikasi_antrian(status);
CREATE INDEX idx_notifikasi_antrian_scheduled_at ON notifikasi_antrian(scheduled_at);

CREATE INDEX idx_audit_log_pengguna_id ON audit_log(pengguna_id);
CREATE INDEX idx_audit_log_aksi ON audit_log(aksi);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);
