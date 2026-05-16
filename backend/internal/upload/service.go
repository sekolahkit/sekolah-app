package upload

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Service struct {
	basePath     string
	maxSize      int64
	allowedTypes []string
}

func NewService(basePath string, maxSizeMB int, allowedTypes []string) *Service {
	return &Service{
		basePath:     basePath,
		maxSize:      int64(maxSizeMB) * 1024 * 1024,
		allowedTypes: allowedTypes,
	}
}

func (s *Service) Upload(sekolahID int64, category string, file multipart.File, header *multipart.FileHeader) (string, error) {
	if header.Size > s.maxSize {
		return "", fmt.Errorf("ukuran file melebihi batas maksimal %d MB", s.maxSize/(1024*1024))
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		contentType = http.DetectContentType(buf[:n])
		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}
	}

	if !s.isAllowedType(contentType) {
		return "", fmt.Errorf("tipe file tidak diizinkan: %s", contentType)
	}

	ext := extensionFromContentType(contentType)
	randomName := generateRandomName() + ext

	dir := filepath.Join(s.basePath, fmt.Sprintf("%d", sekolahID), category)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("buat direktori upload: %w", err)
	}

	filePath := filepath.Join(dir, randomName)
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("buat file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("simpan file: %w", err)
	}

	relativePath := fmt.Sprintf("%d/%s/%s", sekolahID, category, randomName)
	return relativePath, nil
}

func (s *Service) GetFilePath(relativePath string) (string, error) {
	clean := filepath.Clean(relativePath)
	if strings.Contains(clean, "..") {
		return "", fmt.Errorf("path tidak valid")
	}
	fullPath := filepath.Join(s.basePath, clean)
	if _, err := os.Stat(fullPath); err != nil {
		return "", fmt.Errorf("file tidak ditemukan")
	}
	return fullPath, nil
}

func (s *Service) ValidateAccess(relativePath string, sekolahID int64) error {
	parts := strings.SplitN(relativePath, "/", 2)
	if len(parts) < 2 {
		return fmt.Errorf("path tidak valid")
	}
	if parts[0] != fmt.Sprintf("%d", sekolahID) {
		return fmt.Errorf("akses ditolak")
	}
	return nil
}

func (s *Service) isAllowedType(contentType string) bool {
	for _, t := range s.allowedTypes {
		if strings.EqualFold(contentType, t) {
			return true
		}
	}
	return false
}

func generateRandomName() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func extensionFromContentType(ct string) string {
	switch {
	case strings.Contains(ct, "jpeg") || strings.Contains(ct, "jpg"):
		return ".jpg"
	case strings.Contains(ct, "png"):
		return ".png"
	case strings.Contains(ct, "pdf"):
		return ".pdf"
	default:
		return ".bin"
	}
}
