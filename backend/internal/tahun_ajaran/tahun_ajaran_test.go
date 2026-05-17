package tahun_ajaran

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
		INSERT INTO sekolah (id, nama, kode) VALUES (1, 'Test School', 'TS01');
		CREATE TABLE tahun_ajaran (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			nama TEXT NOT NULL,
			aktif BOOLEAN DEFAULT FALSE,
			tanggal_mulai DATE,
			tanggal_selesai DATE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (sekolah_id) REFERENCES sekolah(id)
		);
	`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	ta, err := svc.Create(1, CreateRequest{Nama: "2025/2026"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if ta.Nama != "2025/2026" {
		t.Fatalf("expected nama '2025/2026', got %q", ta.Nama)
	}
	if ta.SekolahID != 1 {
		t.Fatalf("expected sekolah_id 1, got %d", ta.SekolahID)
	}
}

func TestCreateValidation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	_, err := svc.Create(1, CreateRequest{Nama: ""})
	if err == nil {
		t.Fatal("expected validation error for empty nama")
	}
}

func TestList(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	svc.Create(1, CreateRequest{Nama: "2024/2025"})
	svc.Create(1, CreateRequest{Nama: "2025/2026"})

	list, err := svc.List(1)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list))
	}
}

func TestListTenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	db.Exec(`INSERT INTO sekolah (id, nama, kode) VALUES (2, 'Other', 'OT01')`)
	svc.Create(1, CreateRequest{Nama: "2025/2026"})
	svc.Create(2, CreateRequest{Nama: "2025/2026"})

	list, err := svc.List(1)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 item for tenant 1, got %d", len(list))
	}
}

func TestGetAktif(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	svc.Create(1, CreateRequest{Nama: "2024/2025"})
	ta2, _ := svc.Create(1, CreateRequest{Nama: "2025/2026"})

	repo.SetAktif(1, ta2.ID)

	aktif, err := svc.GetAktif(1)
	if err != nil {
		t.Fatalf("GetAktif failed: %v", err)
	}
	if aktif.ID != ta2.ID {
		t.Fatalf("expected aktif ID %d, got %d", ta2.ID, aktif.ID)
	}
	if !aktif.Aktif {
		t.Fatal("expected aktif=true")
	}
}

func TestGetAktifNone(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	_, err := svc.GetAktif(1)
	if err == nil {
		t.Fatal("expected error when no active tahun ajaran")
	}
}

func TestSetAktif(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	ta1, _ := svc.Create(1, CreateRequest{Nama: "2024/2025"})
	ta2, _ := svc.Create(1, CreateRequest{Nama: "2025/2026"})

	if err := svc.SetAktif(1, ta1.ID); err != nil {
		t.Fatalf("SetAktif ta1 failed: %v", err)
	}
	aktif, _ := svc.GetAktif(1)
	if aktif.ID != ta1.ID {
		t.Fatalf("expected ta1 aktif, got %d", aktif.ID)
	}

	if err := svc.SetAktif(1, ta2.ID); err != nil {
		t.Fatalf("SetAktif ta2 failed: %v", err)
	}
	aktif, _ = svc.GetAktif(1)
	if aktif.ID != ta2.ID {
		t.Fatalf("expected ta2 aktif, got %d", aktif.ID)
	}

	list, _ := svc.List(1)
	for _, item := range list {
		if item.ID == ta2.ID && !item.Aktif {
			t.Fatal("ta2 should be aktif")
		}
		if item.ID == ta1.ID && item.Aktif {
			t.Fatal("ta1 should not be aktif after switching")
		}
	}
}

func TestSetAktifNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	err := svc.SetAktif(1, 999)
	if err == nil {
		t.Fatal("expected error for non-existent ID")
	}
}

func TestUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	ta, _ := svc.Create(1, CreateRequest{Nama: "2025/2026"})

	updated, err := svc.Update(1, ta.ID, UpdateRequest{Nama: "2025/2026 Genap", TanggalMulai: "2026-01-15"})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Nama != "2025/2026 Genap" {
		t.Fatalf("expected nama updated, got %q", updated.Nama)
	}
}

func TestUpdateNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	_, err := svc.Update(1, 999, UpdateRequest{Nama: "Test"})
	if err == nil {
		t.Fatal("expected error for non-existent ID")
	}
}
