package ppdb

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE sekolah (id INTEGER PRIMARY KEY AUTOINCREMENT, nama TEXT NOT NULL);
		CREATE TABLE pengguna (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, email TEXT NOT NULL, nama TEXT NOT NULL, password TEXT NOT NULL, role TEXT NOT NULL);
		CREATE TABLE tahun_ajaran (id INTEGER PRIMARY KEY AUTOINCREMENT, sekolah_id INTEGER NOT NULL, nama TEXT NOT NULL, aktif INTEGER DEFAULT 0);
		CREATE TABLE ppdb_pendaftaran (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			tahun_ajaran_id INTEGER NOT NULL,
			nama_lengkap TEXT NOT NULL,
			nik TEXT, tempat_lahir TEXT, tanggal_lahir DATE,
			jenis_kelamin TEXT NOT NULL, agama TEXT, alamat TEXT,
			asal_sekolah TEXT, no_hp TEXT, email TEXT,
			nama_ortu TEXT, no_hp_ortu TEXT, pekerjaan_ortu TEXT,
			foto TEXT, status TEXT DEFAULT 'menunggu',
			skor DECIMAL(10,2), ranking INTEGER,
			latitude REAL, longitude REAL, catatan TEXT,
			daftar_ulang_status TEXT NOT NULL DEFAULT 'belum',
			daftar_ulang_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE ppdb_ujian (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			pendaftaran_id INTEGER NOT NULL,
			nama_ujian TEXT NOT NULL,
			nilai DECIMAL(5,2),
			keterangan TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE ppdb_pengumuman (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			pendaftaran_id INTEGER NOT NULL,
			status TEXT NOT NULL,
			ranking INTEGER, keterangan TEXT,
			tanggal_pengumuman DATE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
			UNIQUE(sekolah_id, tahun_ajaran_id)
		);
		CREATE TABLE ppdb_ranking_log (
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
			executed_at DATETIME DEFAULT (datetime('now'))
		);
		CREATE TABLE ppdb_berkas (id INTEGER PRIMARY KEY AUTOINCREMENT, pendaftaran_id INTEGER NOT NULL, jenis_berkas TEXT NOT NULL, file_path TEXT NOT NULL, status TEXT DEFAULT 'pending', catatan TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP);
	`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func seedTestData(t *testing.T, db *sql.DB) {
	t.Helper()
	db.Exec(`INSERT INTO sekolah (id, nama) VALUES (1, 'Test School')`)
	db.Exec(`INSERT INTO pengguna (id, sekolah_id, email, nama, password, role) VALUES (1, 1, 'admin@test.com', 'Admin', 'hash', 'admin')`)
	db.Exec(`INSERT INTO tahun_ajaran (id, sekolah_id, nama, aktif) VALUES (1, 1, '2024/2025', 1)`)
}

func TestRankingAssignsStatusByKuota(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	seedTestData(t, db)

	repo := NewRepository(db)
	svc := NewService(repo)

	repo.UpsertKonfigurasiRanking(&KonfigurasiRanking{
		SekolahID: 1, TahunAjaranID: 1, Metode: "nilai_ujian", Kuota: 2, Cadangan: 1,
	})

	for i, name := range []string{"Alice", "Bob", "Charlie", "Diana"} {
		repo.CreatePendaftaran(&Pendaftaran{
			SekolahID: 1, TahunAjaranID: 1, NamaLengkap: name,
			JenisKelamin: "L", Status: "menunggu",
		})
		repo.CreateUjian(&Ujian{PendaftaranID: int64(i + 1), NamaUjian: "Test", Nilai: float64(90 - i*10)})
	}

	result, err := svc.RunRanking(1, 1, RunRankingRequest{TahunAjaranID: 1, DryRun: false})
	if err != nil {
		t.Fatalf("RunRanking failed: %v", err)
	}

	if result.DiterimaCount != 2 {
		t.Errorf("expected 2 diterima, got %d", result.DiterimaCount)
	}
	if result.CadanganCount != 1 {
		t.Errorf("expected 1 cadangan, got %d", result.CadanganCount)
	}
	if result.TidakDiterima != 1 {
		t.Errorf("expected 1 tidak_diterima, got %d", result.TidakDiterima)
	}

	if result.Ranked[0].NamaLengkap != "Alice" {
		t.Errorf("expected Alice at rank 1, got %s", result.Ranked[0].NamaLengkap)
	}
	if result.Ranked[0].Status != "diterima" {
		t.Errorf("expected diterima at rank 1, got %s", result.Ranked[0].Status)
	}
	if result.Ranked[2].Status != "cadangan" {
		t.Errorf("expected cadangan at rank 3, got %s", result.Ranked[2].Status)
	}
	if result.Ranked[3].Status != "tidak_diterima" {
		t.Errorf("expected tidak_diterima at rank 4, got %s", result.Ranked[3].Status)
	}
}

func TestRankingIsDeterministic(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	seedTestData(t, db)

	repo := NewRepository(db)
	svc := NewService(repo)

	repo.UpsertKonfigurasiRanking(&KonfigurasiRanking{
		SekolahID: 1, TahunAjaranID: 1, Metode: "nilai_ujian", Kuota: 1, Cadangan: 0,
	})

	repo.CreatePendaftaran(&Pendaftaran{SekolahID: 1, TahunAjaranID: 1, NamaLengkap: "Alice", JenisKelamin: "P", Status: "menunggu", TanggalLahir: "2010-01-01"})
	repo.CreatePendaftaran(&Pendaftaran{SekolahID: 1, TahunAjaranID: 1, NamaLengkap: "Bob", JenisKelamin: "L", Status: "menunggu", TanggalLahir: "2010-02-01"})
	repo.CreateUjian(&Ujian{PendaftaranID: 1, NamaUjian: "Test", Nilai: 80})
	repo.CreateUjian(&Ujian{PendaftaranID: 2, NamaUjian: "Test", Nilai: 80})

	result1, _ := svc.RunRanking(1, 1, RunRankingRequest{TahunAjaranID: 1, DryRun: false})
	result2, _ := svc.RunRanking(1, 1, RunRankingRequest{TahunAjaranID: 1, DryRun: false})

	if result1.Ranked[0].ID != result2.Ranked[0].ID {
		t.Errorf("ranking not deterministic: got %d then %d", result1.Ranked[0].ID, result2.Ranked[0].ID)
	}
}

func TestDryRunDoesNotPersist(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	seedTestData(t, db)

	repo := NewRepository(db)
	svc := NewService(repo)

	repo.UpsertKonfigurasiRanking(&KonfigurasiRanking{
		SekolahID: 1, TahunAjaranID: 1, Metode: "nilai_ujian", Kuota: 1, Cadangan: 0,
	})

	repo.CreatePendaftaran(&Pendaftaran{SekolahID: 1, TahunAjaranID: 1, NamaLengkap: "Alice", JenisKelamin: "P", Status: "menunggu"})
	repo.CreateUjian(&Ujian{PendaftaranID: 1, NamaUjian: "Test", Nilai: 90})

	result, err := svc.RunRanking(1, 1, RunRankingRequest{TahunAjaranID: 1, DryRun: true})
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}
	if !result.DryRun {
		t.Error("expected DryRun=true in result")
	}

	p, _ := repo.GetPendaftaranByID(1, 1)
	if p.Status != "menunggu" {
		t.Errorf("dry run should not change status, got %s", p.Status)
	}
	if p.Ranking != 0 {
		t.Errorf("dry run should not set ranking, got %d", p.Ranking)
	}
}

func TestPersistedRunStoresRankAndScore(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	seedTestData(t, db)

	repo := NewRepository(db)
	svc := NewService(repo)

	repo.UpsertKonfigurasiRanking(&KonfigurasiRanking{
		SekolahID: 1, TahunAjaranID: 1, Metode: "nilai_ujian", Kuota: 1, Cadangan: 0,
	})

	repo.CreatePendaftaran(&Pendaftaran{SekolahID: 1, TahunAjaranID: 1, NamaLengkap: "Alice", JenisKelamin: "P", Status: "menunggu"})
	repo.CreateUjian(&Ujian{PendaftaranID: 1, NamaUjian: "Test", Nilai: 85})

	_, err := svc.RunRanking(1, 1, RunRankingRequest{TahunAjaranID: 1, DryRun: false})
	if err != nil {
		t.Fatalf("RunRanking failed: %v", err)
	}

	p, _ := repo.GetPendaftaranByID(1, 1)
	if p.Ranking != 1 {
		t.Errorf("expected ranking 1, got %d", p.Ranking)
	}
	if p.Skor != 85 {
		t.Errorf("expected skor 85, got %f", p.Skor)
	}
	if p.Status != "diterima" {
		t.Errorf("expected diterima, got %s", p.Status)
	}
}

func TestPublishOnlyAfterRanking(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	seedTestData(t, db)

	repo := NewRepository(db)
	svc := NewService(repo)

	repo.UpsertKonfigurasiRanking(&KonfigurasiRanking{
		SekolahID: 1, TahunAjaranID: 1, Metode: "nilai_ujian", Kuota: 1, Cadangan: 0,
	})

	repo.CreatePendaftaran(&Pendaftaran{SekolahID: 1, TahunAjaranID: 1, NamaLengkap: "Alice", JenisKelamin: "P", Status: "menunggu"})

	_, err := svc.PublishRanking(1, PublishRankingRequest{TahunAjaranID: 1})
	if err == nil {
		t.Fatal("expected error when publishing without ranking")
	}

	repo.CreateUjian(&Ujian{PendaftaranID: 1, NamaUjian: "Test", Nilai: 90})
	svc.RunRanking(1, 1, RunRankingRequest{TahunAjaranID: 1, DryRun: false})

	count, err := svc.PublishRanking(1, PublishRankingRequest{TahunAjaranID: 1})
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 published, got %d", count)
	}
}

func TestDaftarUlangOnlyForDiterima(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	seedTestData(t, db)

	repo := NewRepository(db)
	svc := NewService(repo)

	repo.CreatePendaftaran(&Pendaftaran{SekolahID: 1, TahunAjaranID: 1, NamaLengkap: "Alice", JenisKelamin: "P", Status: "menunggu"})
	err := svc.DaftarUlang(1, 1)
	if err == nil {
		t.Fatal("expected error for non-diterima pendaftar")
	}

	repo.CreatePendaftaran(&Pendaftaran{SekolahID: 1, TahunAjaranID: 1, NamaLengkap: "Bob", JenisKelamin: "L", Status: "diterima"})
	err = svc.DaftarUlang(1, 2)
	if err != nil {
		t.Fatalf("daftar ulang failed: %v", err)
	}

	status, _ := svc.GetDaftarUlangStatus(1, 2)
	if status.DaftarUlangStatus != "belum" {
		t.Errorf("expected belum, got %s", status.DaftarUlangStatus)
	}
}

func TestCrossSchoolRankingRejected(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Exec(`INSERT INTO sekolah (id, nama) VALUES (1, 'School A'), (2, 'School B')`)
	db.Exec(`INSERT INTO pengguna (id, sekolah_id, email, nama, password, role) VALUES (1, 1, 'admin@test.com', 'Admin', 'hash', 'admin')`)
	db.Exec(`INSERT INTO tahun_ajaran (id, sekolah_id, nama, aktif) VALUES (1, 1, '2024/2025', 1)`)

	repo := NewRepository(db)
	svc := NewService(repo)

	repo.UpsertKonfigurasiRanking(&KonfigurasiRanking{
		SekolahID: 2, TahunAjaranID: 1, Metode: "nilai_ujian", Kuota: 1, Cadangan: 0,
	})

	_, err := svc.RunRanking(1, 1, RunRankingRequest{TahunAjaranID: 1})
	if err == nil {
		t.Fatal("expected error for cross-school ranking")
	}
}

func TestReRunRankingReplacesPrevious(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	seedTestData(t, db)

	repo := NewRepository(db)
	svc := NewService(repo)

	repo.UpsertKonfigurasiRanking(&KonfigurasiRanking{
		SekolahID: 1, TahunAjaranID: 1, Metode: "nilai_ujian", Kuota: 1, Cadangan: 0,
	})

	repo.CreatePendaftaran(&Pendaftaran{SekolahID: 1, TahunAjaranID: 1, NamaLengkap: "Alice", JenisKelamin: "P", Status: "menunggu"})
	repo.CreateUjian(&Ujian{PendaftaranID: 1, NamaUjian: "Test", Nilai: 80})

	svc.RunRanking(1, 1, RunRankingRequest{TahunAjaranID: 1, DryRun: false})

	p1, _ := repo.GetPendaftaranByID(1, 1)
	if p1.Status != "diterima" {
		t.Fatalf("expected diterima after first run, got %s", p1.Status)
	}

	repo.UpsertKonfigurasiRanking(&KonfigurasiRanking{
		SekolahID: 1, TahunAjaranID: 1, Metode: "nilai_ujian", Kuota: 0, Cadangan: 0,
	})

	svc.RunRanking(1, 1, RunRankingRequest{TahunAjaranID: 1, DryRun: false})

	p2, _ := repo.GetPendaftaranByID(1, 1)
	if p2.Status != "tidak_diterima" {
		t.Errorf("expected tidak_diterima after re-run with kuota=0, got %s", p2.Status)
	}
}
