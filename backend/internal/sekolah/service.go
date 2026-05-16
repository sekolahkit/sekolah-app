package sekolah

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

type UpdateRequest struct {
	Nama              string `json:"nama"`
	Alamat            string `json:"alamat"`
	Telepon           string `json:"telepon"`
	Email             string `json:"email"`
	Logo              string `json:"logo"`
	Website           string `json:"website"`
	NamaKepalaSekolah string `json:"nama_kepala_sekolah"`
	NIPKepalaSekolah  string `json:"nip_kepala_sekolah"`
}

func (s *Service) Get(sekolahID int64) (*Sekolah, error) {
	return s.repo.GetByID(sekolahID)
}

func (s *Service) Update(sekolahID int64, req UpdateRequest) (*Sekolah, error) {
	errs := validator.Collect(
		validator.Required("nama", req.Nama),
		validator.Email("email", req.Email),
	)
	if len(errs) > 0 {
		return nil, errs
	}

	sek := &Sekolah{
		Nama:              req.Nama,
		Alamat:            req.Alamat,
		Telepon:           req.Telepon,
		Email:             req.Email,
		Logo:              req.Logo,
		Website:           req.Website,
		NamaKepalaSekolah: req.NamaKepalaSekolah,
		NIPKepalaSekolah:  req.NIPKepalaSekolah,
	}

	if err := s.repo.Update(sekolahID, sek); err != nil {
		return nil, fmt.Errorf("update sekolah: %w", err)
	}

	return s.repo.GetByID(sekolahID)
}
