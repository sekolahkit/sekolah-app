package kelas

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
		CREATE TABLE sekolah (id INTEGER PRIMARY KEY AUTOINCREMENT, nama TEXT, kode TEXT);
		INSERT INTO sekolah (id, nama, kode) VALUES (1, 'School A', 'SA'), (2, 'School B', 'SB');

		CREATE TABLE tahun_ajaran (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			nama TEXT NOT NULL,
			aktif BOOLEAN DEFAULT FALSE
		);
		INSERT INTO tahun_ajaran (id, sekolah_id, nama, aktif) VALUES (1, 1, '2025/2026', 1), (2, 2, '2025/2026', 1);

		CREATE TABLE jurusan (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			nama TEXT NOT NULL,
			kode TEXT
		);

		CREATE TABLE pengguna (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			email TEXT NOT NULL,
			password TEXT NOT NULL,
			nama TEXT NOT NULL,
			role TEXT NOT NULL
		);
		INSERT INTO pengguna (id, sekolah_id, email, password, nama, role) VALUES (1, 1, 'guru@test.com', 'x', 'Pak Guru', 'guru');

		CREATE TABLE kelas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			nama TEXT NOT NULL,
			tingkat INTEGER NOT NULL,
			jurusan_id INTEGER,
			wali_kelas_id INTEGER,
			ruangan TEXT,
			kapasitas INTEGER,
			shift TEXT,
			tahun_ajaran_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
			FOREIGN KEY (tahun_ajaran_id) REFERENCES tahun_ajaran(id),
			FOREIGN KEY (wali_kelas_id) REFERENCES pengguna(id)
		);

		CREATE TABLE siswa (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			nis TEXT NOT NULL,
			nama TEXT NOT NULL,
			jenis_kelamin TEXT NOT NULL,
			tahun_ajaran_masuk INTEGER,
			status TEXT DEFAULT 'aktif',
			UNIQUE(sekolah_id, nis)
		);
		INSERT INTO siswa (id, sekolah_id, nis, nama, jenis_kelamin) VALUES (1, 1, '001', 'Siswa A', 'L'), (2, 1, '002', 'Siswa B', 'P'), (3, 2, '001', 'Siswa C', 'L');

		CREATE TABLE kelas_siswa (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			siswa_id INTEGER NOT NULL,
			kelas_id INTEGER NOT NULL,
			tahun_ajaran_id INTEGER NOT NULL,
			status TEXT DEFAULT 'aktif',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (sekolah_id) REFERENCES sekolah(id),
			FOREIGN KEY (siswa_id) REFERENCES siswa(id),
			FOREIGN KEY (kelas_id) REFERENCES kelas(id),
			FOREIGN KEY (tahun_ajaran_id) REFERENCES tahun_ajaran(id),
			UNIQUE(sekolah_id, siswa_id, kelas_id, tahun_ajaran_id)
		);
	`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func createTestKelas(t *testing.T, db *sql.DB, sekolahID int64, nama string, taID int64) int64 {
	t.Helper()
	result, err := db.Exec(`INSERT INTO kelas (sekolah_id, nama, tingkat, tahun_ajaran_id) VALUES (?, ?, 10, ?)`, sekolahID, nama, taID)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := result.LastInsertId()
	return id
}

func TestCreateKelas(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	k, err := svc.Create(1, CreateRequest{TahunAjaranID: 1, Nama: "X IPA 1", Tingkat: "10"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if k.Nama != "X IPA 1" {
		t.Fatalf("expected nama 'X IPA 1', got %q", k.Nama)
	}
	if k.SekolahID != 1 {
		t.Fatalf("expected sekolah_id 1, got %d", k.SekolahID)
	}
	if k.WaliKelasID != nil {
		t.Fatal("expected nil wali_kelas_id")
	}
}

func TestCreateKelasWithWali(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	waliID := int64(1)
	k, err := svc.Create(1, CreateRequest{TahunAjaranID: 1, Nama: "X IPA 1", Tingkat: "10", WaliKelasID: &waliID})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if k.WaliKelasID == nil || *k.WaliKelasID != 1 {
		t.Fatalf("expected wali_kelas_id 1, got %v", k.WaliKelasID)
	}
}

func TestAddSiswaIncludesSekolahAndTahunAjaran(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	kelasID := createTestKelas(t, db, 1, "X IPA 1", 1)

	if err := svc.AddSiswa(1, kelasID, AddSiswaRequest{SiswaID: 1}); err != nil {
		t.Fatalf("AddSiswa failed: %v", err)
	}

	var sekolahID, taID int64
	err := db.QueryRow(`SELECT sekolah_id, tahun_ajaran_id FROM kelas_siswa WHERE kelas_id = ? AND siswa_id = ?`, kelasID, 1).Scan(&sekolahID, &taID)
	if err != nil {
		t.Fatal(err)
	}
	if sekolahID != 1 {
		t.Fatalf("expected sekolah_id 1, got %d", sekolahID)
	}
	if taID != 1 {
		t.Fatalf("expected tahun_ajaran_id 1, got %d", taID)
	}
}

func TestAddSiswaDuplicateIsNoop(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	kelasID := createTestKelas(t, db, 1, "X IPA 1", 1)

	if err := svc.AddSiswa(1, kelasID, AddSiswaRequest{SiswaID: 1}); err != nil {
		t.Fatalf("first AddSiswa failed: %v", err)
	}

	err := svc.AddSiswa(1, kelasID, AddSiswaRequest{SiswaID: 1})
	if err == nil {
		t.Fatal("expected duplicate error")
	}

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM kelas_siswa WHERE kelas_id = ? AND siswa_id = ?`, kelasID, 1).Scan(&count)
	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}
}

func TestAddSiswaFromAnotherSekolah(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	kelasID := createTestKelas(t, db, 1, "X IPA 1", 1)

	err := svc.AddSiswa(1, kelasID, AddSiswaRequest{SiswaID: 3})
	if err == nil {
		t.Fatal("expected error for siswa from another sekolah")
	}
}

func TestAddSiswaToKelasFromAnotherSekolah(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	kelasID := createTestKelas(t, db, 1, "X IPA 1", 1)

	err := svc.AddSiswa(2, kelasID, AddSiswaRequest{SiswaID: 3})
	if err == nil {
		t.Fatal("expected error for kelas from another sekolah")
	}
}

func TestRemoveSiswa(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	kelasID := createTestKelas(t, db, 1, "X IPA 1", 1)
	svc.AddSiswa(1, kelasID, AddSiswaRequest{SiswaID: 1})

	if err := svc.RemoveSiswa(1, kelasID, 1); err != nil {
		t.Fatalf("RemoveSiswa failed: %v", err)
	}

	ids, _ := svc.ListSiswa(1, kelasID)
	if len(ids) != 0 {
		t.Fatalf("expected 0 siswa after remove, got %d", len(ids))
	}
}

func TestRemoveSiswaFromAnotherSekolah(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	kelasID := createTestKelas(t, db, 1, "X IPA 1", 1)
	svc.AddSiswa(1, kelasID, AddSiswaRequest{SiswaID: 1})

	err := svc.RemoveSiswa(2, kelasID, 1)
	if err == nil {
		t.Fatal("expected error for removing from another sekolah")
	}
}

func TestListSiswa(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	kelasID := createTestKelas(t, db, 1, "X IPA 1", 1)
	svc.AddSiswa(1, kelasID, AddSiswaRequest{SiswaID: 1})
	svc.AddSiswa(1, kelasID, AddSiswaRequest{SiswaID: 2})

	ids, err := svc.ListSiswa(1, kelasID)
	if err != nil {
		t.Fatalf("ListSiswa failed: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 siswa, got %d", len(ids))
	}
}

func TestListSiswaTenantScoped(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	kelasID := createTestKelas(t, db, 1, "X IPA 1", 1)
	svc.AddSiswa(1, kelasID, AddSiswaRequest{SiswaID: 1})

	ids, err := svc.ListSiswa(2, kelasID)
	if err == nil {
		t.Fatal("expected error for kelas from another sekolah")
	}
	if ids != nil {
		t.Fatalf("expected nil ids, got %v", ids)
	}
}

func TestListKelasTenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	createTestKelas(t, db, 1, "X IPA 1", 1)
	createTestKelas(t, db, 2, "X IPA 2", 2)

	list, total, err := svc.List(1, ListParams{Page: 1, Limit: 10, Sort: "created_at", Order: "desc"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected 1 kelas for sekolah 1, got %d", total)
	}
	if list[0].Nama != "X IPA 1" {
		t.Fatalf("expected 'X IPA 1', got %q", list[0].Nama)
	}
}

func TestGetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	waliID := int64(1)
	k, _ := svc.Create(1, CreateRequest{TahunAjaranID: 1, Nama: "X IPA 1", Tingkat: "10", WaliKelasID: &waliID})

	got, err := svc.GetByID(1, k.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.WaliKelasID == nil || *got.WaliKelasID != 1 {
		t.Fatalf("expected wali_kelas_id 1, got %v", got.WaliKelasID)
	}
}

func TestGetByIDTenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	k, _ := svc.Create(1, CreateRequest{TahunAjaranID: 1, Nama: "X IPA 1", Tingkat: "10"})

	_, err := svc.GetByID(2, k.ID)
	if err == nil {
		t.Fatal("expected error for kelas from another sekolah")
	}
}

func TestAddSiswaRepoLevelOnConflict(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	kelasID := createTestKelas(t, db, 1, "X IPA 1", 1)

	if err := repo.AddSiswa(1, kelasID, 1, 1); err != nil {
		t.Fatalf("first insert failed: %v", err)
	}

	if err := repo.AddSiswa(1, kelasID, 1, 1); err != nil {
		t.Fatalf("second insert should be silent no-op, got: %v", err)
	}

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM kelas_siswa WHERE sekolah_id = 1 AND kelas_id = ? AND siswa_id = 1 AND tahun_ajaran_id = 1`, kelasID).Scan(&count)
	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}
}

func TestSameIDsAcrossSekolahNoCollision(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)

	kelasID1 := createTestKelas(t, db, 1, "X IPA 1", 1)
	kelasID2 := createTestKelas(t, db, 2, "X IPA 1", 2)

	if err := repo.AddSiswa(1, kelasID1, 1, 1); err != nil {
		t.Fatalf("insert sekolah 1 failed: %v", err)
	}
	if err := repo.AddSiswa(2, kelasID2, 3, 2); err != nil {
		t.Fatalf("insert sekolah 2 failed: %v", err)
	}

	var count1, count2 int
	db.QueryRow(`SELECT COUNT(*) FROM kelas_siswa WHERE sekolah_id = 1`).Scan(&count1)
	db.QueryRow(`SELECT COUNT(*) FROM kelas_siswa WHERE sekolah_id = 2`).Scan(&count2)
	if count1 != 1 || count2 != 1 {
		t.Fatalf("expected 1 row per sekolah, got %d and %d", count1, count2)
	}
}

func TestSchemaUniqueIndexMatchesMigration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	var sql string
	err := db.QueryRow(`SELECT sql FROM sqlite_master WHERE name = 'kelas_siswa' AND type = 'table'`).Scan(&sql)
	if err != nil {
		t.Fatal(err)
	}

	if !contains(sql, "UNIQUE(sekolah_id, siswa_id, kelas_id, tahun_ajaran_id)") {
		t.Fatalf("expected UNIQUE constraint with sekolah_id, got: %s", sql)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
