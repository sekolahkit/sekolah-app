CREATE TABLE IF NOT EXISTS payment_callback_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    provider TEXT NOT NULL,
    payment_gateway_id TEXT NOT NULL,
    order_id TEXT NOT NULL,
    status TEXT NOT NULL,
    amount INTEGER NOT NULL DEFAULT 0,
    sekolah_id INTEGER NOT NULL,
    processed INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT (datetime('now')),
    UNIQUE(provider, payment_gateway_id)
);
