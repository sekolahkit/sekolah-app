ALTER TABLE notifikasi_antrian ADD COLUMN claimed_at DATETIME;
CREATE INDEX idx_notifikasi_antrian_claimed_at ON notifikasi_antrian(claimed_at);
