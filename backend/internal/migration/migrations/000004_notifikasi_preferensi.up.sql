CREATE TABLE notifikasi_preferensi (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    pengguna_id INTEGER,
    siswa_id INTEGER,
    recipient_type TEXT NOT NULL DEFAULT 'manual',
    channel TEXT NOT NULL,
    destination TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    consent_status TEXT NOT NULL DEFAULT 'pending',
    consent_source TEXT NOT NULL DEFAULT 'admin',
    consent_at DATETIME,
    revoked_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    FOREIGN KEY (pengguna_id) REFERENCES pengguna(id),
    FOREIGN KEY (siswa_id) REFERENCES siswa(id)
);

CREATE UNIQUE INDEX idx_notifikasi_preferensi_unique
    ON notifikasi_preferensi(sekolah_id, channel, destination);

CREATE INDEX idx_notifikasi_preferensi_sekolah
    ON notifikasi_preferensi(sekolah_id);

CREATE INDEX idx_notifikasi_preferensi_lookup
    ON notifikasi_preferensi(sekolah_id, channel, destination, consent_status, enabled);
