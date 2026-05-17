package backup

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupTestEnv(t *testing.T) (Config, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	backupPath := filepath.Join(tmpDir, "backups")
	dbPath := filepath.Join(tmpDir, "data", "sekolah.db")
	uploadPath := filepath.Join(tmpDir, "uploads")

	os.MkdirAll(filepath.Dir(dbPath), 0755)
	os.MkdirAll(filepath.Join(uploadPath, "1", "bukti_bayar"), 0755)
	os.MkdirAll(backupPath, 0755)

	dbContent := []byte("SQLite format 3\ntest database content")
	os.WriteFile(dbPath, dbContent, 0644)

	uploadContent := []byte("test upload file content")
	os.WriteFile(filepath.Join(uploadPath, "1", "bukti_bayar", "test.jpg"), uploadContent, 0644)
	os.WriteFile(filepath.Join(uploadPath, "1", "bukti_bayar", "test2.jpg"), uploadContent, 0644)

	cfg := Config{
		BackupPath: backupPath,
		DBPath:     dbPath,
		UploadPath: uploadPath,
		Retention:  7,
	}

	return cfg, func() {
		os.RemoveAll(tmpDir)
	}
}

func TestCreate_BackupCreated(t *testing.T) {
	cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	svc := NewService(cfg)
	info, err := svc.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if info.ID == "" {
		t.Error("backup ID kosong")
	}
	if info.Filename == "" {
		t.Error("backup filename kosong")
	}
	if info.Size == 0 {
		t.Error("backup size 0")
	}
	if info.Checksum == "" {
		t.Error("backup checksum kosong")
	}

	backupFile := filepath.Join(cfg.BackupPath, info.Filename)
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Error("file backup tidak ditemukan")
	}
}

func TestCreate_ArchiveContainsDB(t *testing.T) {
	cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	svc := NewService(cfg)
	info, err := svc.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	backupFile := filepath.Join(cfg.BackupPath, info.Filename)
	files := listTarGzContents(t, backupFile)

	foundDB := false
	foundMetadata := false
	for _, f := range files {
		if f == dbFilename {
			foundDB = true
		}
		if f == metadataFilename {
			foundMetadata = true
		}
	}

	if !foundDB {
		t.Error("database file tidak ditemukan dalam archive")
	}
	if !foundMetadata {
		t.Error("metadata file tidak ditemukan dalam archive")
	}
}

func TestCreate_ArchiveContainsUploads(t *testing.T) {
	cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	svc := NewService(cfg)
	info, err := svc.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	backupFile := filepath.Join(cfg.BackupPath, info.Filename)
	files := listTarGzContents(t, backupFile)

	foundUpload := false
	for _, f := range files {
		if strings.Contains(f, "bukti_bayar") {
			foundUpload = true
			break
		}
	}

	if !foundUpload {
		t.Error("upload files tidak ditemukan dalam archive")
	}
}

func TestList_ReturnsBackups(t *testing.T) {
	cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	svc := NewService(cfg)

	svc.Create()
	time.Sleep(1100 * time.Millisecond)
	svc.Create()

	backups, err := svc.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(backups) < 2 {
		t.Errorf("List() returned %d backups, want >= 2", len(backups))
	}

	for _, b := range backups {
		if b.ID == "" {
			t.Error("backup ID kosong")
		}
		if b.Filename == "" {
			t.Error("backup filename kosong")
		}
	}
}

func TestList_EmptyDirectory(t *testing.T) {
	cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	svc := NewService(cfg)

	backups, err := svc.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(backups) != 0 {
		t.Errorf("List() returned %d backups, want 0", len(backups))
	}
}

func TestGetByID_Exists(t *testing.T) {
	cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	svc := NewService(cfg)
	created, err := svc.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := svc.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("GetByID() ID = %s, want %s", found.ID, created.ID)
	}
}

func TestGetByID_NotExists(t *testing.T) {
	cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	svc := NewService(cfg)

	_, err := svc.GetByID("nonexistent")
	if err == nil {
		t.Error("GetByID() expected error for nonexistent backup")
	}
}

func TestGetFilePath_Valid(t *testing.T) {
	cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	svc := NewService(cfg)
	created, _ := svc.Create()

	filePath, filename, err := svc.GetFilePath(created.ID)
	if err != nil {
		t.Fatalf("GetFilePath() error = %v", err)
	}

	if filePath == "" {
		t.Error("file path kosong")
	}
	if filename == "" {
		t.Error("filename kosong")
	}
}

func TestGetFilePath_PathTraversal(t *testing.T) {
	cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	svc := NewService(cfg)

	_, _, err := svc.GetFilePath("../../etc/passwd")
	if err == nil {
		t.Error("GetFilePath() expected error for path traversal")
	}
}

func TestRestore_Success(t *testing.T) {
	cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	svc := NewService(cfg)
	created, err := svc.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	backupFile := filepath.Join(cfg.BackupPath, created.Filename)
	backupCopy := filepath.Join(cfg.BackupPath, "backup_copy.tar.gz")
	copyFile(backupFile, backupCopy)

	newDBContent := []byte("modified database")
	os.WriteFile(cfg.DBPath, newDBContent, 0644)

	time.Sleep(1100 * time.Millisecond)
	err = svc.Restore(created.ID)
	if err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	dbContent, err := os.ReadFile(cfg.DBPath)
	if err != nil {
		t.Fatalf("reading restored DB: %v", err)
	}

	if string(dbContent) != "SQLite format 3\ntest database content" {
		t.Errorf("database content tidak sesuai setelah restore, got: %s", string(dbContent))
	}
}

func TestRestore_CreatesPreRestoreBackup(t *testing.T) {
	cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	svc := NewService(cfg)
	created, _ := svc.Create()

	backupsBefore, _ := svc.List()
	countBefore := len(backupsBefore)

	time.Sleep(1100 * time.Millisecond)
	err := svc.Restore(created.ID)
	if err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	backupsAfter, _ := svc.List()
	countAfter := len(backupsAfter)

	if countAfter <= countBefore {
		t.Error("pre-restore backup tidak dibuat")
	}
}

func TestRestore_NonexistentBackup(t *testing.T) {
	cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	svc := NewService(cfg)

	err := svc.Restore("nonexistent")
	if err == nil {
		t.Error("Restore() expected error for nonexistent backup")
	}
}

func TestCleanup_RespectsRetention(t *testing.T) {
	cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	cfg.Retention = 0
	svc := NewService(cfg)

	svc.Create()

	err := svc.Cleanup()
	if err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	backups, _ := svc.List()
	if len(backups) != 1 {
		t.Errorf("Cleanup() with retention=0 should not delete, got %d backups", len(backups))
	}
}

func TestValidateArchive_ValidFile(t *testing.T) {
	cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	svc := NewService(cfg)
	created, _ := svc.Create()

	backupPath := filepath.Join(cfg.BackupPath, created.Filename)
	err := svc.validateArchive(backupPath)
	if err != nil {
		t.Fatalf("validateArchive() error = %v", err)
	}
}

func TestValidateArchive_InvalidFile(t *testing.T) {
	cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	svc := NewService(cfg)

	invalidFile := filepath.Join(cfg.BackupPath, "invalid.tar.gz")
	os.WriteFile(invalidFile, []byte("not a valid archive"), 0644)

	err := svc.validateArchive(invalidFile)
	if err == nil {
		t.Error("validateArchive() expected error for invalid file")
	}
}

func TestValidateArchive_PathTraversal(t *testing.T) {
	cfg, cleanup := setupTestEnv(t)
	defer cleanup()

	svc := NewService(cfg)

	maliciousPath := filepath.Join(cfg.BackupPath, "..", "..", "evil.tar.gz")
	err := svc.validateArchive(maliciousPath)
	if err == nil {
		t.Error("validateArchive() expected error for path traversal")
	}
}

func listTarGzContents(t *testing.T, filePath string) []string {
	t.Helper()

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	var names []string
	for {
		header, err := tr.Next()
		if err != nil {
			break
		}
		names = append(names, header.Name)
	}
	return names
}
