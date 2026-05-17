package selfservice

import (
	"database/sql"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreatePembayaranRequest struct {
	TagihanID         int64   `json:"tagihan_id"`
	Jumlah            float64 `json:"jumlah"`
	Tanggal           string  `json:"tanggal"`
	Metode            string  `json:"metode"`
	BuktiBayar        string  `json:"bukti_bayar"`
	RekeningSekolahID int64   `json:"rekening_sekolah_id"`
	Catatan           string  `json:"catatan"`
}

func (s *Service) GetLinkedSiswa(sekolahID, penggunaID int64) ([]LinkedSiswa, error) {
	return s.repo.GetLinkedSiswa(sekolahID, penggunaID)
}

func (s *Service) GetSiswaDetail(sekolahID, penggunaID, siswaID int64) (*SiswaDetail, error) {
	ok, err := s.repo.HasAccess(sekolahID, penggunaID, siswaID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return s.repo.GetSiswaDetail(sekolahID, siswaID)
}

func (s *Service) GetTagihan(sekolahID, penggunaID, siswaID int64) ([]Tagihan, error) {
	ok, err := s.repo.HasAccess(sekolahID, penggunaID, siswaID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return s.repo.GetTagihanBySiswa(sekolahID, siswaID)
}

func (s *Service) GetPembayaran(sekolahID, penggunaID, siswaID int64) ([]Pembayaran, error) {
	ok, err := s.repo.HasAccess(sekolahID, penggunaID, siswaID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return s.repo.GetPembayaranBySiswa(sekolahID, siswaID)
}

func (s *Service) CreatePembayaran(sekolahID, penggunaID int64, req CreatePembayaranRequest) (int64, error) {
	ok, siswaID, err := s.repo.TagihanBelongsToLinkedSiswa(sekolahID, penggunaID, req.TagihanID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("not found")
		}
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("not found")
	}

	return s.repo.CreatePembayaran(req.TagihanID, siswaID, req.Jumlah, req.Tanggal, req.Metode, req.BuktiBayar, req.RekeningSekolahID, req.Catatan)
}

func (s *Service) DashboardSiswa(sekolahID, penggunaID int64) (*DashboardSiswa, error) {
	linked, err := s.repo.GetLinkedSiswa(sekolahID, penggunaID)
	if err != nil {
		return nil, err
	}
	if len(linked) == 0 {
		return &DashboardSiswa{}, nil
	}
	return s.repo.GetDashboardSiswa(sekolahID, linked[0].ID)
}

func (s *Service) DashboardOrangtua(sekolahID, penggunaID int64) (*DashboardOrangtua, error) {
	return s.repo.GetDashboardOrangtua(sekolahID, penggunaID)
}

func (s *Service) DashboardGuru(sekolahID, penggunaID int64) (*DashboardGuru, error) {
	return s.repo.GetDashboardGuru(sekolahID, penggunaID)
}

func (s *Service) GetGuruKelas(sekolahID, penggunaID int64) ([]GuruKelas, error) {
	return s.repo.GetGuruKelas(sekolahID, penggunaID)
}

func (s *Service) GetGuruSiswaByKelas(sekolahID, penggunaID, kelasID int64) ([]GuruSiswa, error) {
	result, err := s.repo.GetGuruSiswaByKelas(sekolahID, penggunaID, kelasID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("not found")
		}
		return nil, err
	}
	return result, nil
}
