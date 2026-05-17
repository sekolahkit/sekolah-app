package jurusan

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
		CREATE TABLE jurusan (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			nama TEXT NOT NULL,
			kode TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (sekolah_id) REFERENCES sekolah(id)
		);
		CREATE TABLE kelas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sekolah_id INTEGER NOT NULL,
			nama TEXT NOT NULL,
			tingkat INTEGER NOT NULL,
			jurusan_id INTEGER,
			tahun_ajaran_id INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
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

	j, err := svc.Create(1, CreateRequest{Nama: "Teknik Informatika", Kode: "TI"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if j.Nama != "Teknik Informatika" {
		t.Fatalf("expected nama 'Teknik Informatika', got %q", j.Nama)
	}
	if j.Kode != "TI" {
		t.Fatalf("expected kode 'TI', got %q", j.Kode)
	}
}

func TestCreateValidation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	_, err := svc.Create(1, CreateRequest{Nama: "", Kode: "TI"})
	if err == nil {
		t.Fatal("expected validation error for empty nama")
	}
}

func TestList(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	svc.Create(1, CreateRequest{Nama: "Teknik Informatika", Kode: "TI"})
	svc.Create(1, CreateRequest{Nama: "Akuntansi", Kode: "AK"})

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
	svc.Create(1, CreateRequest{Nama: "TI", Kode: "TI"})
	svc.Create(2, CreateRequest{Nama: "TI", Kode: "TI"})

	list, err := svc.List(1)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 item for tenant 1, got %d", len(list))
	}
}

func TestUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	j, _ := svc.Create(1, CreateRequest{Nama: "TI", Kode: "TI"})

	updated, err := svc.Update(1, j.ID, UpdateRequest{Nama: "Teknik Informatika", Kode: "TI01"})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Nama != "Teknik Informatika" {
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

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	j, _ := svc.Create(1, CreateRequest{Nama: "TI", Kode: "TI"})

	if err := svc.Delete(1, j.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	list, _ := svc.List(1)
	if len(list) != 0 {
		t.Fatalf("expected 0 items after delete, got %d", len(list))
	}
}

func TestDeleteNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	err := svc.Delete(1, 999)
	if err == nil {
		t.Fatal("expected error for non-existent ID")
	}
}

func TestDeleteProtectedWhenReferenced(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	j, _ := svc.Create(1, CreateRequest{Nama: "TI", Kode: "TI"})

	_, err := db.Exec(`INSERT INTO kelas (sekolah_id, nama, tingkat, jurusan_id, tahun_ajaran_id) VALUES (1, 'TI-1', 3, ?, 1)`, j.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = svc.Delete(1, j.ID)
	if err == nil {
		t.Fatal("expected error when deleting referenced jurusan")
	}

	list, _ := svc.List(1)
	if len(list) != 1 {
		t.Fatalf("expected jurusan to still exist, got %d items", len(list))
	}
}

func TestDeleteWhenNotReferenced(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	j, _ := svc.Create(1, CreateRequest{Nama: "TI", Kode: "TI"})
	svc.Create(1, CreateRequest{Nama: "AK", Kode: "AK"})

	if err := svc.Delete(1, j.ID); err != nil {
		t.Fatalf("Delete should succeed when not referenced: %v", err)
	}

	list, _ := svc.List(1)
	if len(list) != 1 {
		t.Fatalf("expected 1 remaining, got %d", len(list))
	}
}
