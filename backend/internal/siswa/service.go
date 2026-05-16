package siswa

import (
	"fmt"

	"github.com/Sekolahkit/sekolah-app/pkg/validator"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreateRequest struct {
	NIS          string `json:"nis"`
	Nama         string `json:"nama"`
	JenisKelamin string `json:"jenis_kelamin"`
	TempatLahir  string `json:"tempat_lahir"`
	TanggalLahir string `json:"tanggal_lahir"`
	Agama        string `json:"agama"`
	Alamat       string `json:"alamat"`
	NoHP         string `json:"no_hp"`
	Email        string `json:"email"`
	NamaOrtu     string `json:"nama_ortu"`
	NoHPOrtu     string `json:"no_hp_ortu"`
	EmailOrtu    string `json:"email_ortu"`
}

type UpdateRequest struct {
	NIS          string `json:"nis"`
	Nama         string `json:"nama"`
	JenisKelamin string `json:"jenis_kelamin"`
	TempatLahir  string `json:"tempat_lahir"`
	TanggalLahir string `json:"tanggal_lahir"`
	Agama        string `json:"agama"`
	Alamat       string `json:"alamat"`
	NoHP         string `json:"no_hp"`
	Email        string `json:"email"`
	NamaOrtu     string `json:"nama_ortu"`
	NoHPOrtu     string `json:"no_hp_ortu"`
	EmailOrtu    string `json:"email_ortu"`
}

func (s *Service) List(sekolahID int64, params ListParams) ([]Siswa, int, error) {
	return s.repo.List(sekolahID, params)
}

func (s *Service) GetByID(sekolahID, id int64) (*Siswa, error) {
	return s.repo.GetByID(sekolahID, id)
}

func (s *Service) Create(sekolahID int64, req CreateRequest) (*Siswa, error) {
	errs := validator.Collect(
		validator.Required("nis", req.NIS),
		validator.Required("nama", req.Nama),
		validator.Required("jenis_kelamin", req.JenisKelamin),
		validator.InList("jenis_kelamin", req.JenisKelamin, []string{"L", "P"}),
	)
	if len(errs) > 0 {
		return nil, errs
	}

	exists, err := s.repo.NISExists(sekolahID, req.NIS, 0)
	if err != nil {
		return nil, fmt.Errorf("cek NIS: %w", err)
	}
	if exists {
		return nil, validator.ValidationErrors{{Field: "nis", Message: "NIS sudah digunakan"}}
	}

	siswa := &Siswa{
		SekolahID:    sekolahID,
		NIS:          req.NIS,
		Nama:         req.Nama,
		JenisKelamin: req.JenisKelamin,
		TempatLahir:  req.TempatLahir,
		TanggalLahir: req.TanggalLahir,
		Agama:        req.Agama,
		Alamat:       req.Alamat,
		NoHP:         req.NoHP,
		Email:        req.Email,
		NamaOrtu:     req.NamaOrtu,
		NoHPOrtu:     req.NoHPOrtu,
		EmailOrtu:    req.EmailOrtu,
	}

	id, err := s.repo.Create(siswa)
	if err != nil {
		return nil, fmt.Errorf("create siswa: %w", err)
	}

	return s.repo.GetByID(sekolahID, id)
}

func (s *Service) Update(sekolahID, id int64, req UpdateRequest) (*Siswa, error) {
	errs := validator.Collect(
		validator.Required("nis", req.NIS),
		validator.Required("nama", req.Nama),
		validator.Required("jenis_kelamin", req.JenisKelamin),
		validator.InList("jenis_kelamin", req.JenisKelamin, []string{"L", "P"}),
	)
	if len(errs) > 0 {
		return nil, errs
	}

	exists, err := s.repo.NISExists(sekolahID, req.NIS, id)
	if err != nil {
		return nil, fmt.Errorf("cek NIS: %w", err)
	}
	if exists {
		return nil, validator.ValidationErrors{{Field: "nis", Message: "NIS sudah digunakan"}}
	}

	siswa := &Siswa{
		NIS:          req.NIS,
		Nama:         req.Nama,
		JenisKelamin: req.JenisKelamin,
		TempatLahir:  req.TempatLahir,
		TanggalLahir: req.TanggalLahir,
		Agama:        req.Agama,
		Alamat:       req.Alamat,
		NoHP:         req.NoHP,
		Email:        req.Email,
		NamaOrtu:     req.NamaOrtu,
		NoHPOrtu:     req.NoHPOrtu,
		EmailOrtu:    req.EmailOrtu,
	}

	if err := s.repo.Update(sekolahID, id, siswa); err != nil {
		return nil, fmt.Errorf("update siswa: %w", err)
	}

	return s.repo.GetByID(sekolahID, id)
}

func (s *Service) Delete(sekolahID, id int64) error {
	_, err := s.repo.GetByID(sekolahID, id)
	if err != nil {
		return fmt.Errorf("siswa tidak ditemukan")
	}
	return s.repo.Delete(sekolahID, id)
}
