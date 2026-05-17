package backup

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	backupFormatVersion = 1
	metadataFilename    = "metadata.json"
	dbFilename          = "sekolah.db"
	uploadsDirname      = "uploads"
	preRestorePrefix    = "pre-restore_"
)

type Metadata struct {
	CreatedAt       string `json:"created_at"`
	FormatVersion   int    `json:"format_version"`
	DatabaseFile    string `json:"database_file"`
	UploadFileCount int    `json:"upload_file_count"`
	Checksum        string `json:"checksum"`
	AppVersion      string `json:"app_version,omitempty"`
	FileSize        int64  `json:"file_size"`
}

type BackupInfo struct {
	ID        string `json:"id"`
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
	Checksum  string `json:"checksum"`
}

type Config struct {
	BackupPath string
	DBPath     string
	UploadPath string
	Retention  int
}

type Service struct {
	cfg Config
}

func NewService(cfg Config) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) Create() (*BackupInfo, error) {
	if err := os.MkdirAll(s.cfg.BackupPath, 0755); err != nil {
		return nil, fmt.Errorf("membuat direktori backup: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("backup_%s.tar.gz", timestamp)
	backupPath := filepath.Join(s.cfg.BackupPath, filename)

	dbFileCount, err := countFiles(s.cfg.UploadPath)
	if err != nil {
		slog.Warn("menghitung file upload", "error", err)
		dbFileCount = 0
	}

	if err := s.createArchive(backupPath); err != nil {
		os.Remove(backupPath)
		return nil, err
	}

	checksum, err := computeChecksum(backupPath)
	if err != nil {
		os.Remove(backupPath)
		return nil, fmt.Errorf("menghitung checksum: %w", err)
	}

	stat, err := os.Stat(backupPath)
	if err != nil {
		os.Remove(backupPath)
		return nil, fmt.Errorf("membaca info file: %w", err)
	}

	info := &BackupInfo{
		ID:        timestamp,
		Filename:  filename,
		Size:      stat.Size(),
		CreatedAt: time.Now().Format(time.RFC3339),
		Checksum:  checksum,
	}

	_ = dbFileCount

	slog.Info("backup dibuat", "filename", filename, "size", stat.Size())
	return info, nil
}

func (s *Service) List() ([]BackupInfo, error) {
	entries, err := os.ReadDir(s.cfg.BackupPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupInfo{}, nil
		}
		return nil, fmt.Errorf("membaca direktori backup: %w", err)
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tar.gz") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}

		checksum, _ := computeChecksum(filepath.Join(s.cfg.BackupPath, entry.Name()))

		id := strings.TrimPrefix(entry.Name(), "backup_")
		id = strings.TrimSuffix(id, ".tar.gz")

		backups = append(backups, BackupInfo{
			ID:        id,
			Filename:  entry.Name(),
			Size:      info.Size(),
			CreatedAt: info.ModTime().Format(time.RFC3339),
			Checksum:  checksum,
		})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt > backups[j].CreatedAt
	})

	return backups, nil
}

func (s *Service) GetByID(id string) (*BackupInfo, error) {
	filename := fmt.Sprintf("backup_%s.tar.gz", id)
	backupPath := filepath.Join(s.cfg.BackupPath, filename)

	stat, err := os.Stat(backupPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("backup tidak ditemukan")
		}
		return nil, fmt.Errorf("membaca file backup: %w", err)
	}

	checksum, _ := computeChecksum(backupPath)

	return &BackupInfo{
		ID:        id,
		Filename:  filename,
		Size:      stat.Size(),
		CreatedAt: stat.ModTime().Format(time.RFC3339),
		Checksum:  checksum,
	}, nil
}

func (s *Service) GetFilePath(id string) (string, string, error) {
	filename := fmt.Sprintf("backup_%s.tar.gz", id)
	backupPath := filepath.Join(s.cfg.BackupPath, filename)

	cleanPath := filepath.Clean(backupPath)
	if !strings.HasPrefix(cleanPath, filepath.Clean(s.cfg.BackupPath)) {
		return "", "", fmt.Errorf("path tidak valid")
	}

	if _, err := os.Stat(cleanPath); err != nil {
		return "", "", fmt.Errorf("backup tidak ditemukan")
	}

	return cleanPath, filename, nil
}

func (s *Service) Restore(id string) error {
	backupPath := filepath.Join(s.cfg.BackupPath, fmt.Sprintf("backup_%s.tar.gz", id))
	cleanPath := filepath.Clean(backupPath)
	if !strings.HasPrefix(cleanPath, filepath.Clean(s.cfg.BackupPath)) {
		return fmt.Errorf("path tidak valid")
	}

	if _, err := os.Stat(cleanPath); err != nil {
		return fmt.Errorf("backup tidak ditemukan: %w", err)
	}

	if err := s.validateArchive(cleanPath); err != nil {
		return fmt.Errorf("archive tidak valid: %w", err)
	}

	if _, err := s.Create(); err != nil {
		slog.Warn("gagal membuat pre-restore backup", "error", err)
	}

	tmpDir, err := os.MkdirTemp("", "restore_*")
	if err != nil {
		return fmt.Errorf("membuat direktori temporary: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := s.extractArchive(cleanPath, tmpDir); err != nil {
		return fmt.Errorf("mengekstrak archive: %w", err)
	}

	tmpDB := filepath.Join(tmpDir, dbFilename)
	if _, err := os.Stat(tmpDB); err != nil {
		return fmt.Errorf("database file tidak ditemukan dalam backup")
	}

	tmpUploads := filepath.Join(tmpDir, uploadsDirname)

	dbDir := filepath.Dir(s.cfg.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("membuat direktori database: %w", err)
	}

	if err := copyFile(tmpDB, s.cfg.DBPath); err != nil {
		return fmt.Errorf("mengganti database: %w", err)
	}

	if _, err := os.Stat(tmpUploads); err == nil {
		if err := os.RemoveAll(s.cfg.UploadPath); err != nil {
			return fmt.Errorf("menghapus uploads lama: %w", err)
		}
		if err := copyDir(tmpUploads, s.cfg.UploadPath); err != nil {
			return fmt.Errorf("mengganti uploads: %w", err)
		}
	}

	slog.Info("restore selesai", "backup_id", id)
	return nil
}

func (s *Service) Cleanup() error {
	if s.cfg.Retention <= 0 {
		return nil
	}

	backups, err := s.List()
	if err != nil {
		return err
	}

	cutoff := time.Now().AddDate(0, 0, -s.cfg.Retention)
	deleted := 0

	for _, b := range backups {
		t, err := time.Parse(time.RFC3339, b.CreatedAt)
		if err != nil {
			continue
		}
		if t.Before(cutoff) {
			path := filepath.Join(s.cfg.BackupPath, b.Filename)
			if err := os.Remove(path); err == nil {
				deleted++
			}
		}
	}

	if deleted > 0 {
		slog.Info("retention cleanup", "deleted", deleted)
	}
	return nil
}

func (s *Service) createArchive(archivePath string) error {
	file, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("membuat file archive: %w", err)
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	if _, err := os.Stat(s.cfg.DBPath); err == nil {
		if err := addFileToTar(tw, s.cfg.DBPath, dbFilename); err != nil {
			return fmt.Errorf("menambahkan database ke archive: %w", err)
		}
	}

	if _, err := os.Stat(s.cfg.UploadPath); err == nil {
		baseLen := len(filepath.Dir(s.cfg.UploadPath)) + 1
		err := filepath.Walk(s.cfg.UploadPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			relPath := path[baseLen:]
			return addFileToTar(tw, path, relPath)
		})
		if err != nil {
			return fmt.Errorf("menambahkan uploads ke archive: %w", err)
		}
	}

	uploadCount, _ := countFiles(s.cfg.UploadPath)
	metadata := Metadata{
		CreatedAt:       time.Now().Format(time.RFC3339),
		FormatVersion:   backupFormatVersion,
		DatabaseFile:    dbFilename,
		UploadFileCount: uploadCount,
		AppVersion:      "0.1.0",
	}

	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("membuat metadata: %w", err)
	}

	header := &tar.Header{
		Name: metadataFilename,
		Mode: 0644,
		Size: int64(len(metadataJSON)),
	}
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("menulis header metadata: %w", err)
	}
	if _, err := tw.Write(metadataJSON); err != nil {
		return fmt.Errorf("menulis metadata: %w", err)
	}

	return nil
}

func (s *Service) validateArchive(archivePath string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("membuka archive: %w", err)
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("membuka gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	foundDB := false
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("membaca archive: %w", err)
		}

		cleanName := filepath.Clean(header.Name)
		if strings.Contains(cleanName, "..") {
			return fmt.Errorf("path traversal terdeteksi: %s", header.Name)
		}

		if cleanName == dbFilename {
			foundDB = true
		}
	}

	if !foundDB {
		return fmt.Errorf("database file tidak ditemukan dalam archive")
	}

	return nil
}

func (s *Service) extractArchive(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		cleanName := filepath.Clean(header.Name)
		if strings.Contains(cleanName, "..") {
			continue
		}

		target := filepath.Join(destDir, cleanName)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			continue
		}

		if header.Typeflag == tar.TypeDir {
			os.MkdirAll(target, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	return nil
}

func addFileToTar(tw *tar.Writer, filePath, archiveName string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name: archiveName,
		Mode: 0644,
		Size: stat.Size(),
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err = io.Copy(tw, file)
	return err
}

func computeChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func countFiles(dir string) (int, error) {
	if dir == "" {
		return 0, nil
	}
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count, err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		target := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		return copyFile(path, target)
	})
}
