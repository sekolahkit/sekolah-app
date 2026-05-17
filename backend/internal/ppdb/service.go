package ppdb

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

type DaftarRequest struct {
	TahunAjaranID int64   `json:"tahun_ajaran_id"`
	NamaLengkap   string  `json:"nama_lengkap"`
	NIK           string  `json:"nik"`
	TempatLahir   string  `json:"tempat_lahir"`
	TanggalLahir  string  `json:"tanggal_lahir"`
	JenisKelamin  string  `json:"jenis_kelamin"`
	Agama         string  `json:"agama"`
	Alamat        string  `json:"alamat"`
	AsalSekolah   string  `json:"asal_sekolah"`
	NoHP          string  `json:"no_hp"`
	Email         string  `json:"email"`
	NamaOrtu      string  `json:"nama_ortu"`
	NoHPOrtu      string  `json:"no_hp_ortu"`
	PekerjaanOrtu string  `json:"pekerjaan_ortu"`
	Foto          string  `json:"foto"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

type VerifikasiBerkasRequest struct {
	Status  string `json:"status"`
	Catatan string `json:"catatan"`
}

type InputUjianRequest struct {
	PendaftaranID int64   `json:"pendaftaran_id"`
	NamaUjian     string  `json:"nama_ujian"`
	Nilai         float64 `json:"nilai"`
	Keterangan    string  `json:"keterangan"`
}

type PublishPengumumanRequest struct {
	PendaftaranID int64  `json:"pendaftaran_id"`
	Status        string `json:"status"`
	Ranking       int    `json:"ranking"`
	Keterangan    string `json:"keterangan"`
}

type KonfigurasiRankingRequest struct {
	TahunAjaranID int64  `json:"tahun_ajaran_id"`
	Metode        string `json:"metode"`
	BobotJSON     string `json:"bobot_json"`
	Kuota         int    `json:"kuota"`
	Cadangan      int    `json:"cadangan"`
}

var allowedStatuses = []string{"menunggu", "berkas_lengkap", "berkas_ditolak", "diterima", "cadangan", "tidak_diterima", "daftar_ulang"}

func (s *Service) Daftar(req DaftarRequest) (*Pendaftaran, error) {
	errs := validator.Collect(
		validator.Required("nama_lengkap", req.NamaLengkap),
		validator.Required("jenis_kelamin", req.JenisKelamin),
		validator.InList("jenis_kelamin", req.JenisKelamin, []string{"L", "P"}),
	)
	if req.TahunAjaranID == 0 {
		errs = append(errs, validator.ValidationError{Field: "tahun_ajaran_id", Message: "wajib diisi"})
	}
	if len(errs) > 0 {
		return nil, errs
	}

	sekolahID, err := s.repo.GetSekolahIDByTahunAjaran(req.TahunAjaranID)
	if err != nil {
		return nil, fmt.Errorf("tahun ajaran tidak ditemukan")
	}

	p := &Pendaftaran{
		SekolahID:     sekolahID,
		TahunAjaranID: req.TahunAjaranID,
		NamaLengkap:   req.NamaLengkap,
		NIK:           req.NIK,
		TempatLahir:   req.TempatLahir,
		TanggalLahir:  req.TanggalLahir,
		JenisKelamin:  req.JenisKelamin,
		Agama:         req.Agama,
		Alamat:        req.Alamat,
		AsalSekolah:   req.AsalSekolah,
		NoHP:          req.NoHP,
		Email:         req.Email,
		NamaOrtu:      req.NamaOrtu,
		NoHPOrtu:      req.NoHPOrtu,
		PekerjaanOrtu: req.PekerjaanOrtu,
		Foto:          req.Foto,
		Status:        "menunggu",
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
	}

	id, err := s.repo.CreatePendaftaran(p)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pendaftaran: %w", err)
	}

	return s.repo.GetPendaftaranByID(sekolahID, id)
}

func (s *Service) ListPendaftar(sekolahID int64, params ListParams) ([]Pendaftaran, int, error) {
	return s.repo.ListPendaftaran(sekolahID, params)
}

func (s *Service) GetPendaftar(sekolahID, id int64) (*Pendaftaran, error) {
	return s.repo.GetPendaftaranByID(sekolahID, id)
}

func (s *Service) UpdateStatus(sekolahID, id int64, req UpdateStatusRequest) error {
	errs := validator.Collect(
		validator.Required("status", req.Status),
		validator.InList("status", req.Status, allowedStatuses),
	)
	if len(errs) > 0 {
		return errs
	}

	_, err := s.repo.GetPendaftaranByID(sekolahID, id)
	if err != nil {
		return fmt.Errorf("pendaftaran tidak ditemukan")
	}

	return s.repo.UpdatePendaftaranStatus(sekolahID, id, req.Status)
}

func (s *Service) ListBerkas(sekolahID, pendaftaranID int64) ([]Berkas, error) {
	_, err := s.repo.GetPendaftaranByID(sekolahID, pendaftaranID)
	if err != nil {
		return nil, fmt.Errorf("pendaftaran tidak ditemukan")
	}
	return s.repo.ListBerkasByPendaftaran(pendaftaranID)
}

func (s *Service) VerifikasiBerkas(sekolahID, id int64, req VerifikasiBerkasRequest) error {
	errs := validator.Collect(
		validator.Required("status", req.Status),
		validator.InList("status", req.Status, []string{"pending", "diterima", "ditolak"}),
	)
	if len(errs) > 0 {
		return errs
	}
	return s.repo.UpdateBerkasStatus(sekolahID, id, req.Status, req.Catatan)
}

func (s *Service) InputNilaiUjian(sekolahID int64, req InputUjianRequest) (*Ujian, error) {
	errs := validator.Collect(
		validator.Required("nama_ujian", req.NamaUjian),
	)
	if req.PendaftaranID == 0 {
		errs = append(errs, validator.ValidationError{Field: "pendaftaran_id", Message: "wajib diisi"})
	}
	if len(errs) > 0 {
		return nil, errs
	}

	_, err := s.repo.GetPendaftaranByID(sekolahID, req.PendaftaranID)
	if err != nil {
		return nil, fmt.Errorf("pendaftaran tidak ditemukan")
	}

	u := &Ujian{
		PendaftaranID: req.PendaftaranID,
		NamaUjian:     req.NamaUjian,
		Nilai:         req.Nilai,
		Keterangan:    req.Keterangan,
	}

	id, err := s.repo.CreateUjian(u)
	if err != nil {
		return nil, fmt.Errorf("gagal input nilai ujian: %w", err)
	}

	u.ID = id
	return u, nil
}

func (s *Service) PublishPengumuman(sekolahID int64, req PublishPengumumanRequest) (*Pengumuman, error) {
	errs := validator.Collect(
		validator.Required("status", req.Status),
	)
	if req.PendaftaranID == 0 {
		errs = append(errs, validator.ValidationError{Field: "pendaftaran_id", Message: "wajib diisi"})
	}
	if len(errs) > 0 {
		return nil, errs
	}

	_, err := s.repo.GetPendaftaranByID(sekolahID, req.PendaftaranID)
	if err != nil {
		return nil, fmt.Errorf("pendaftaran tidak ditemukan")
	}

	p := &Pengumuman{
		PendaftaranID: req.PendaftaranID,
		Status:        req.Status,
		Ranking:       req.Ranking,
		Keterangan:    req.Keterangan,
	}

	id, err := s.repo.CreatePengumuman(p)
	if err != nil {
		return nil, fmt.Errorf("gagal publish pengumuman: %w", err)
	}

	p.ID = id
	return p, nil
}

func (s *Service) GetKonfigurasiRanking(sekolahID, tahunAjaranID int64) (*KonfigurasiRanking, error) {
	return s.repo.GetKonfigurasiRanking(sekolahID, tahunAjaranID)
}

func (s *Service) UpsertKonfigurasiRanking(sekolahID int64, req KonfigurasiRankingRequest) error {
	errs := validator.Collect(
		validator.Required("metode", req.Metode),
	)
	if req.TahunAjaranID == 0 {
		errs = append(errs, validator.ValidationError{Field: "tahun_ajaran_id", Message: "wajib diisi"})
	}
	if req.Kuota <= 0 {
		errs = append(errs, validator.ValidationError{Field: "kuota", Message: "harus lebih dari 0"})
	}
	if len(errs) > 0 {
		return errs
	}

	k := &KonfigurasiRanking{
		SekolahID:     sekolahID,
		TahunAjaranID: req.TahunAjaranID,
		Metode:        req.Metode,
		BobotJSON:     req.BobotJSON,
		Kuota:         req.Kuota,
		Cadangan:      req.Cadangan,
	}

	return s.repo.UpsertKonfigurasiRanking(k)
}

func (s *Service) GetPengumuman(pendaftaranID int64) (*Pengumuman, error) {
	return s.repo.GetPengumumanByPendaftaranID(pendaftaranID)
}

func (s *Service) ExportPendaftar(sekolahID int64, params ListParams) ([]Pendaftaran, error) {
	return s.repo.ListAllPendaftaran(sekolahID, params)
}
