CREATE TABLE telegram_invite (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    preference_id INTEGER NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    invite_expires_at DATETIME NOT NULL,
    invite_used_at DATETIME,
    telegram_user_id INTEGER,
    telegram_chat_id INTEGER,
    telegram_username TEXT,
    status TEXT NOT NULL DEFAULT 'pending_invite',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    FOREIGN KEY (preference_id) REFERENCES notifikasi_preferensi(id)
);

CREATE INDEX idx_telegram_invite_hash ON telegram_invite(token_hash);
CREATE INDEX idx_telegram_invite_pref ON telegram_invite(preference_id);
CREATE INDEX idx_telegram_invite_chat ON telegram_invite(telegram_chat_id);
