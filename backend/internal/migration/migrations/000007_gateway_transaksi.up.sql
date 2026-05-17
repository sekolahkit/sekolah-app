CREATE TABLE IF NOT EXISTS gateway_transaksi (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sekolah_id INTEGER NOT NULL,
    tagihan_id INTEGER NOT NULL,
    provider TEXT NOT NULL,
    order_id TEXT NOT NULL,
    payment_gateway_id TEXT,
    payment_url TEXT NOT NULL,
    amount INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    expires_at DATETIME,
    created_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT (datetime('now')),
    updated_at DATETIME DEFAULT (datetime('now')),
    FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
    FOREIGN KEY (tagihan_id) REFERENCES tagihan(id),
    FOREIGN KEY (created_by) REFERENCES pengguna(id),
    UNIQUE(tagihan_id, provider, status)
);

CREATE INDEX IF NOT EXISTS idx_gateway_transaksi_order_id ON gateway_transaksi(order_id);
CREATE INDEX IF NOT EXISTS idx_gateway_transaksi_sekolah ON gateway_transaksi(sekolah_id);
