package rekening

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
	NamaBank      string `json:"nama_bank"`
	NomorRekening string `json:"nomor_rekening"`
	NamaPemilik   string `json:"nama_pemilik"`
	Cabang        string `json:"cabang"`
	Urutan        int    `json:"urutan"`
	Catatan       string `json:"catatan"`
}

type UpdateRequest struct {
	NamaBank      string `json:"nama_bank"`
	NomorRekening string `json:"nomor_rekening"`
	NamaPemilik   string `json:"nama_pemilik"`
	Cabang        string `json:"cabang"`
	Urutan        int    `json:"urutan"`
	Catatan       string `json:"catatan"`
}

func (s *Service) List(sekolahID int64, params ListParams) ([]Rekening, int, error) {
	return s.repo.List(sekolahID, params)
}

func (s *Service) GetByID(sekolahID, id int64) (*Rekening, error) {
	return s.repo.GetByID(sekolahID, id)
}

func (s *Service) ListAktif(sekolahID int64) ([]Rekening, error) {
	return s.repo.ListAktif(sekolahID)
}

func (s *Service) Create(sekolahID int64, req CreateRequest) (*Rekening, error) {
	errs := validator.Collect(
		validator.Required("nama_bank", req.NamaBank),
		validator.Required("nomor_rekening", req.NomorRekening),
		validator.Required("nama_pemilik", req.NamaPemilik),
	)
	if len(errs) > 0 {
		return nil, errs
	}

	rek := &Rekening{
		SekolahID:     sekolahID,
		NamaBank:      req.NamaBank,
		NomorRekening: req.NomorRekening,
		NamaPemilik:   req.NamaPemilik,
		Cabang:        req.Cabang,
		Urutan:        req.Urutan,
		Catatan:       req.Catatan,
	}

	id, err := s.repo.Create(rek)
	if err != nil {
		return nil, fmt.Errorf("create rekening: %w", err)
	}

	return s.repo.GetByID(sekolahID, id)
}

func (s *Service) Update(sekolahID, id int64, req UpdateRequest) (*Rekening, error) {
	errs := validator.Collect(
		validator.Required("nama_bank", req.NamaBank),
		validator.Required("nomor_rekening", req.NomorRekening),
		validator.Required("nama_pemilik", req.NamaPemilik),
	)
	if len(errs) > 0 {
		return nil, errs
	}

	rek := &Rekening{
		NamaBank:      req.NamaBank,
		NomorRekening: req.NomorRekening,
		NamaPemilik:   req.NamaPemilik,
		Cabang:        req.Cabang,
		Urutan:        req.Urutan,
		Catatan:       req.Catatan,
	}

	if err := s.repo.Update(sekolahID, id, rek); err != nil {
		return nil, fmt.Errorf("update rekening: %w", err)
	}

	return s.repo.GetByID(sekolahID, id)
}

func (s *Service) Delete(sekolahID, id int64) error {
	_, err := s.repo.GetByID(sekolahID, id)
	if err != nil {
		return fmt.Errorf("rekening tidak ditemukan")
	}
	return s.repo.SoftDelete(sekolahID, id)
}
