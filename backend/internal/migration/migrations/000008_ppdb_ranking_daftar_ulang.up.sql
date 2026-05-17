ALTER TABLE ppdb_pendaftaran ADD COLUMN daftar_ulang_status TEXT NOT NULL DEFAULT 'belum';
ALTER TABLE ppdb_pendaftaran ADD COLUMN daftar_ulang_at DATETIME;

CREATE TABLE IF NOT EXISTS ppdb_ranking_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    tahun_ajaran_id INTEGER NOT NULL,
    metode TEXT NOT NULL,
    bobot_json TEXT,
    kuota INTEGER NOT NULL,
    cadangan INTEGER NOT NULL DEFAULT 0,
    total_pendaftar INTEGER NOT NULL DEFAULT 0,
    diterima_count INTEGER NOT NULL DEFAULT 0,
    cadangan_count INTEGER NOT NULL DEFAULT 0,
    tidak_diterima_count INTEGER NOT NULL DEFAULT 0,
    dry_run INTEGER NOT NULL DEFAULT 0,
    executed_by INTEGER NOT NULL,
    executed_at DATETIME DEFAULT (datetime('now')),
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    FOREIGN KEY (tahun_ajaran_id) REFERENCES tahun_ajaran(id),
    FOREIGN KEY (executed_by) REFERENCES pengguna(id)
);

CREATE INDEX IF NOT EXISTS idx_ranking_log_sekolah_ta ON ppdb_ranking_log(sekolah_id, tahun_ajaran_id);
CREATE INDEX IF NOT EXISTS idx_pendaftaran_daftar_ulang ON ppdb_pendaftaran(daftar_ulang_status);
