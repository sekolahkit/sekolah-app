package ppdb

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

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

type RunRankingRequest struct {
	TahunAjaranID int64 `json:"tahun_ajaran_id"`
	DryRun        bool  `json:"dry_run"`
}

type RankingResult struct {
	Ranked        []RankedPendaftaran `json:"ranked"`
	TotalPendaftar int                `json:"total_pendaftar"`
	DiterimaCount  int                `json:"diterima_count"`
	CadanganCount  int                `json:"cadangan_count"`
	TidakDiterima  int                `json:"tidak_diterima_count"`
	Metode         string             `json:"metode"`
	Kuota          int                `json:"kuota"`
	Cadangan       int                `json:"cadangan"`
	DryRun         bool               `json:"dry_run"`
}

func (s *Service) RunRanking(sekolahID, executedBy int64, req RunRankingRequest) (*RankingResult, error) {
	if req.TahunAjaranID == 0 {
		return nil, fmt.Errorf("tahun_ajaran_id wajib diisi")
	}

	konfig, err := s.repo.GetKonfigurasiRanking(sekolahID, req.TahunAjaranID)
	if err != nil {
		return nil, fmt.Errorf("konfigurasi ranking belum diatur untuk tahun ajaran ini")
	}

	pendaftarans, err := s.repo.GetAllPendaftaranForRanking(sekolahID, req.TahunAjaranID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data pendaftar: %w", err)
	}

	type scored struct {
		Pendaftaran Pendaftaran
		Skor        float64
	}

	var items []scored
	for _, p := range pendaftarans {
		var skor float64

		switch konfig.Metode {
		case "nilai_ujian":
			avg, err := s.repo.GetUjianAvgByPendaftaran(p.ID)
			if err != nil {
				avg = 0
			}
			skor = avg

		case "umur":
			if p.TanggalLahir != "" {
				t, err := time.Parse("2006-01-02", p.TanggalLahir)
				if err == nil {
					age := time.Since(t).Hours() / 24 / 365.25
					skor = 100 - age
				}
			}

		case "jarak":
			if p.Latitude != 0 && p.Longitude != 0 {
				dist := haversine(p.Latitude, p.Longitude, 0, 0)
				skor = 100 - dist
				if skor < 0 {
					skor = 0
				}
			}

		case "kombinasi":
			skor = s.computeKombinasiSkor(p, konfig)

		default:
			avg, err := s.repo.GetUjianAvgByPendaftaran(p.ID)
			if err != nil {
				avg = 0
			}
			skor = avg
		}

		items = append(items, scored{Pendaftaran: p, Skor: skor})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Skor != items[j].Skor {
			return items[i].Skor > items[j].Skor
		}
		if items[i].Pendaftaran.TanggalLahir != items[j].Pendaftaran.TanggalLahir {
			return items[i].Pendaftaran.TanggalLahir < items[j].Pendaftaran.TanggalLahir
		}
		return items[i].Pendaftaran.ID < items[j].Pendaftaran.ID
	})

	result := &RankingResult{
		Metode:         konfig.Metode,
		Kuota:          konfig.Kuota,
		Cadangan:       konfig.Cadangan,
		TotalPendaftar: len(items),
		DryRun:         req.DryRun,
	}

	var ranked []RankedPendaftaran
	for i, item := range items {
		rank := i + 1
		var status string
		if rank <= konfig.Kuota {
			status = "diterima"
			result.DiterimaCount++
		} else if rank <= konfig.Kuota+konfig.Cadangan {
			status = "cadangan"
			result.CadanganCount++
		} else {
			status = "tidak_diterima"
			result.TidakDiterima++
		}

		ranked = append(ranked, RankedPendaftaran{
			ID:           item.Pendaftaran.ID,
			NamaLengkap:  item.Pendaftaran.NamaLengkap,
			Skor:         math.Round(item.Skor*100) / 100,
			Ranking:      rank,
			Status:       status,
			TanggalLahir: item.Pendaftaran.TanggalLahir,
			Latitude:     item.Pendaftaran.Latitude,
			Longitude:    item.Pendaftaran.Longitude,
		})
	}
	result.Ranked = ranked

	if !req.DryRun {
		if err := s.repo.ResetRankingForTahunAjaran(sekolahID, req.TahunAjaranID); err != nil {
			return nil, fmt.Errorf("gagal reset ranking: %w", err)
		}

		for _, r := range ranked {
			if err := s.repo.UpdatePendaftaranSkorRankingStatus(sekolahID, r.ID, r.Skor, r.Ranking, r.Status); err != nil {
				return nil, fmt.Errorf("gagal update ranking pendaftar %d: %w", r.ID, err)
			}
		}

		bobotJSON := ""
		if konfig.BobotJSON != "" {
			bobotJSON = konfig.BobotJSON
		}

		log := &RankingLog{
			SekolahID:          sekolahID,
			TahunAjaranID:      req.TahunAjaranID,
			Metode:             konfig.Metode,
			BobotJSON:          bobotJSON,
			Kuota:              konfig.Kuota,
			Cadangan:           konfig.Cadangan,
			TotalPendaftar:     result.TotalPendaftar,
			DiterimaCount:      result.DiterimaCount,
			CadanganCount:      result.CadanganCount,
			TidakDiterimaCount: result.TidakDiterima,
			DryRun:             false,
			ExecutedBy:         executedBy,
		}
		if _, err := s.repo.CreateRankingLog(log); err != nil {
			return nil, fmt.Errorf("gagal menyimpan log ranking: %w", err)
		}
	}

	return result, nil
}

func (s *Service) computeKombinasiSkor(p Pendaftaran, konfig *KonfigurasiRanking) float64 {
	type bobot struct {
		NilaiUjian float64 `json:"nilai_ujian"`
		Umur       float64 `json:"umur"`
		Jarak      float64 `json:"jarak"`
	}
	var b bobot
	if konfig.BobotJSON != "" {
		json.Unmarshal([]byte(konfig.BobotJSON), &b)
	}

	totalBobot := b.NilaiUjian + b.Umur + b.Jarak
	if totalBobot == 0 {
		avg, _ := s.repo.GetUjianAvgByPendaftaran(p.ID)
		return avg
	}

	var skor float64

	if b.NilaiUjian > 0 {
		avg, _ := s.repo.GetUjianAvgByPendaftaran(p.ID)
		skor += avg * (b.NilaiUjian / totalBobot)
	}

	if b.Umur > 0 && p.TanggalLahir != "" {
		t, err := time.Parse("2006-01-02", p.TanggalLahir)
		if err == nil {
			age := time.Since(t).Hours() / 24 / 365.25
			skor += (100 - age) * (b.Umur / totalBobot)
		}
	}

	if b.Jarak > 0 && p.Latitude != 0 && p.Longitude != 0 {
		dist := haversine(p.Latitude, p.Longitude, 0, 0)
		jarakSkor := 100 - dist
		if jarakSkor < 0 {
			jarakSkor = 0
		}
		skor += jarakSkor * (b.Jarak / totalBobot)
	}

	return skor
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

type PublishRankingRequest struct {
	TahunAjaranID int64  `json:"tahun_ajaran_id"`
	Keterangan    string `json:"keterangan"`
}

func (s *Service) PublishRanking(sekolahID int64, req PublishRankingRequest) (int, error) {
	if req.TahunAjaranID == 0 {
		return 0, fmt.Errorf("tahun_ajaran_id wajib diisi")
	}

	pendaftarans, err := s.repo.GetAllPendaftaranForRanking(sekolahID, req.TahunAjaranID)
	if err != nil {
		return 0, fmt.Errorf("gagal mengambil data pendaftar: %w", err)
	}

	hasRanking := false
	for _, p := range pendaftarans {
		if p.Ranking > 0 {
			hasRanking = true
			break
		}
	}
	if !hasRanking {
		return 0, fmt.Errorf("ranking belum dijalankan untuk tahun ajaran ini")
	}

	var pengumumanList []Pengumuman
	now := time.Now().Format("2006-01-02")
	for _, p := range pendaftarans {
		if p.Ranking > 0 {
			ket := req.Keterangan
			if ket == "" {
				switch p.Status {
				case "diterima":
					ket = "Selamat, Anda diterima!"
				case "cadangan":
					ket = "Anda masuk daftar cadangan."
				default:
					ket = "Mohon maaf, Anda belum diterima."
				}
			}
			pengumumanList = append(pengumumanList, Pengumuman{
				PendaftaranID:     p.ID,
				Status:            p.Status,
				Ranking:           p.Ranking,
				Keterangan:        ket,
				TanggalPengumuman: now,
			})
		}
	}

	if len(pengumumanList) == 0 {
		return 0, fmt.Errorf("tidak ada data untuk dipublish")
	}

	if err := s.repo.CreateBulkPengumuman(pengumumanList); err != nil {
		return 0, fmt.Errorf("gagal publish pengumuman: %w", err)
	}

	return len(pengumumanList), nil
}

func (s *Service) DaftarUlang(sekolahID, pendaftaranID int64) error {
	p, err := s.repo.GetPendaftaranByID(sekolahID, pendaftaranID)
	if err != nil {
		return fmt.Errorf("pendaftaran tidak ditemukan")
	}

	if p.Status != "diterima" {
		return fmt.Errorf("hanya pendaftar yang diterima yang dapat daftar ulang")
	}

	return s.repo.UpdateDaftarUlangStatus(sekolahID, pendaftaranID, "sudah")
}

type DaftarUlangStatus struct {
	PendaftaranID    int64  `json:"pendaftaran_id"`
	NamaLengkap      string `json:"nama_lengkap"`
	Status           string `json:"status"`
	DaftarUlangStatus string `json:"daftar_ulang_status"`
	DaftarUlangAt    string `json:"daftar_ulang_at"`
	Ranking          int    `json:"ranking"`
	Skor             float64 `json:"skor"`
}

func (s *Service) GetDaftarUlangStatus(sekolahID, pendaftaranID int64) (*DaftarUlangStatus, error) {
	p, err := s.repo.GetPendaftaranByID(sekolahID, pendaftaranID)
	if err != nil {
		return nil, fmt.Errorf("pendaftaran tidak ditemukan")
	}

	return &DaftarUlangStatus{
		PendaftaranID:    p.ID,
		NamaLengkap:      p.NamaLengkap,
		Status:           p.Status,
		DaftarUlangStatus: "belum",
		DaftarUlangAt:    "",
		Ranking:          p.Ranking,
		Skor:             p.Skor,
	}, nil
}
