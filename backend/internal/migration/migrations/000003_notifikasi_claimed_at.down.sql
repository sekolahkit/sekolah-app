DROP INDEX IF EXISTS idx_notifikasi_antrian_claimed_at;
-- SQLite does not support DROP COLUMN in all versions, so we leave the column in place for the down migration.
