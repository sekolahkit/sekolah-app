package kelas

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
	TahunAjaranID int64  `json:"tahun_ajaran_id"`
	Nama          string `json:"nama"`
	Tingkat       string `json:"tingkat"`
	JurusanID     int64  `json:"jurusan_id"`
	WaliKelasID   *int64 `json:"wali_kelas_id"`
}

type UpdateRequest struct {
	TahunAjaranID int64  `json:"tahun_ajaran_id"`
	Nama          string `json:"nama"`
	Tingkat       string `json:"tingkat"`
	JurusanID     int64  `json:"jurusan_id"`
	WaliKelasID   *int64 `json:"wali_kelas_id"`
}

type AddSiswaRequest struct {
	SiswaID int64 `json:"siswa_id"`
}

func (s *Service) List(sekolahID int64, params ListParams) ([]Kelas, int, error) {
	return s.repo.List(sekolahID, params)
}

func (s *Service) GetByID(sekolahID, id int64) (*Kelas, error) {
	return s.repo.GetByID(sekolahID, id)
}

func (s *Service) Create(sekolahID int64, req CreateRequest) (*Kelas, error) {
	errs := validator.Collect(
		validator.Required("nama", req.Nama),
		validator.Required("tingkat", req.Tingkat),
	)
	if len(errs) > 0 {
		return nil, errs
	}

	if req.TahunAjaranID == 0 {
		return nil, validator.ValidationErrors{{Field: "tahun_ajaran_id", Message: "wajib diisi"}}
	}

	k := &Kelas{
		SekolahID:     sekolahID,
		TahunAjaranID: req.TahunAjaranID,
		Nama:          req.Nama,
		Tingkat:       req.Tingkat,
		JurusanID:     req.JurusanID,
		WaliKelasID:   req.WaliKelasID,
	}

	id, err := s.repo.Create(k)
	if err != nil {
		return nil, fmt.Errorf("create kelas: %w", err)
	}

	return s.repo.GetByID(sekolahID, id)
}

func (s *Service) Update(sekolahID, id int64, req UpdateRequest) (*Kelas, error) {
	errs := validator.Collect(
		validator.Required("nama", req.Nama),
		validator.Required("tingkat", req.Tingkat),
	)
	if len(errs) > 0 {
		return nil, errs
	}

	if req.TahunAjaranID == 0 {
		return nil, validator.ValidationErrors{{Field: "tahun_ajaran_id", Message: "wajib diisi"}}
	}

	k := &Kelas{
		TahunAjaranID: req.TahunAjaranID,
		Nama:          req.Nama,
		Tingkat:       req.Tingkat,
		JurusanID:     req.JurusanID,
		WaliKelasID:   req.WaliKelasID,
	}

	if err := s.repo.Update(sekolahID, id, k); err != nil {
		return nil, fmt.Errorf("update kelas: %w", err)
	}

	return s.repo.GetByID(sekolahID, id)
}

func (s *Service) Delete(sekolahID, id int64) error {
	_, err := s.repo.GetByID(sekolahID, id)
	if err != nil {
		return fmt.Errorf("kelas tidak ditemukan")
	}
	return s.repo.Delete(sekolahID, id)
}

func (s *Service) AddSiswa(sekolahID, kelasID int64, req AddSiswaRequest) error {
	if req.SiswaID == 0 {
		return validator.ValidationErrors{{Field: "siswa_id", Message: "wajib diisi"}}
	}

	kelas, err := s.repo.GetByID(sekolahID, kelasID)
	if err != nil {
		return fmt.Errorf("kelas tidak ditemukan")
	}

	if !s.repo.SiswaExistsInSekolah(sekolahID, req.SiswaID) {
		return fmt.Errorf("siswa tidak ditemukan")
	}

	exists, err := s.repo.SiswaInKelas(sekolahID, kelasID, req.SiswaID)
	if err != nil {
		return fmt.Errorf("cek siswa: %w", err)
	}
	if exists {
		return validator.ValidationErrors{{Field: "siswa_id", Message: "siswa sudah ada di kelas ini"}}
	}

	return s.repo.AddSiswa(sekolahID, kelasID, req.SiswaID, kelas.TahunAjaranID)
}

func (s *Service) RemoveSiswa(sekolahID, kelasID, siswaID int64) error {
	_, err := s.repo.GetByID(sekolahID, kelasID)
	if err != nil {
		return fmt.Errorf("kelas tidak ditemukan")
	}

	exists, err := s.repo.SiswaInKelas(sekolahID, kelasID, siswaID)
	if err != nil {
		return fmt.Errorf("cek siswa: %w", err)
	}
	if !exists {
		return fmt.Errorf("siswa tidak ada di kelas ini")
	}

	return s.repo.RemoveSiswa(sekolahID, kelasID, siswaID)
}

func (s *Service) ListSiswa(sekolahID, kelasID int64) ([]int64, error) {
	_, err := s.repo.GetByID(sekolahID, kelasID)
	if err != nil {
		return nil, fmt.Errorf("kelas tidak ditemukan")
	}
	return s.repo.ListSiswa(sekolahID, kelasID)
}
