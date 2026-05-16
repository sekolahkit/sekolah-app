package setup

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type SetupRequest struct {
	NamaSekolah    string `json:"nama_sekolah"`
	KodeSekolah    string `json:"kode_sekolah"`
	AlamatSekolah  string `json:"alamat_sekolah"`
	TeleponSekolah string `json:"telepon_sekolah"`
	EmailSekolah   string `json:"email_sekolah"`
	NamaAdmin      string `json:"nama_admin"`
	EmailAdmin     string `json:"email_admin"`
	PasswordAdmin  string `json:"password_admin"`
}

type SetupResponse struct {
	SekolahID int64  `json:"sekolah_id"`
	AdminID   int64  `json:"admin_id"`
	Kode      string `json:"kode"`
}

func (s *Service) IsInitialized() (bool, error) {
	return s.repo.IsInitialized()
}

func (s *Service) RunSetup(req SetupRequest) (*SetupResponse, error) {
	initialized, err := s.repo.IsInitialized()
	if err != nil {
		return nil, fmt.Errorf("check initialized: %w", err)
	}
	if initialized {
		return nil, fmt.Errorf("aplikasi sudah di-setup")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.PasswordAdmin), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	sekolahID, err := s.repo.CreateSekolah(
		req.NamaSekolah,
		req.KodeSekolah,
		req.AlamatSekolah,
		req.TeleponSekolah,
		req.EmailSekolah,
	)
	if err != nil {
		return nil, fmt.Errorf("create sekolah: %w", err)
	}

	adminID, err := s.repo.CreateAdmin(sekolahID, req.EmailAdmin, string(hash), req.NamaAdmin)
	if err != nil {
		return nil, fmt.Errorf("create admin: %w", err)
	}

	return &SetupResponse{
		SekolahID: sekolahID,
		AdminID:   adminID,
		Kode:      req.KodeSekolah,
	}, nil
}
