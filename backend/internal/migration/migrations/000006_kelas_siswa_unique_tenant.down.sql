DROP INDEX IF EXISTS idx_kelas_siswa_unique;
CREATE UNIQUE INDEX idx_kelas_siswa_unique ON kelas_siswa(siswa_id, kelas_id, tahun_ajaran_id);
