DROP TABLE IF EXISTS ppdb_ranking_log;

-- SQLite does not support DROP COLUMN before 3.35.0, so we leave the columns.
-- If needed, recreate the table without daftar_ulang_status and daftar_ulang_at.
