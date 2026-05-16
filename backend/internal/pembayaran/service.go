package pembayaran

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

type CreateKategoriRequest struct {
	Nama      string `json:"nama"`
	Deskripsi string `json:"deskripsi"`
}

type UpdateKategoriRequest struct {
	Nama      string `json:"nama"`
	Deskripsi string `json:"deskripsi"`
}

type CreateTagihanRequest struct {
	SiswaID       int64   `json:"siswa_id"`
	KategoriID    int64   `json:"kategori_id"`
	TahunAjaranID int64   `json:"tahun_ajaran_id"`
	Semester      string  `json:"semester"`
	Nominal       float64 `json:"nominal"`
	JatuhTempo    string  `json:"jatuh_tempo"`
	Catatan       string  `json:"catatan"`
}

type BulkCreateTagihanRequest struct {
	SiswaIDs      []int64 `json:"siswa_ids"`
	KategoriID    int64   `json:"kategori_id"`
	TahunAjaranID int64   `json:"tahun_ajaran_id"`
	Semester      string  `json:"semester"`
	Nominal       float64 `json:"nominal"`
	JatuhTempo    string  `json:"jatuh_tempo"`
	Catatan       string  `json:"catatan"`
}

type UpdateTagihanRequest struct {
	KategoriID    int64   `json:"kategori_id"`
	TahunAjaranID int64   `json:"tahun_ajaran_id"`
	Semester      string  `json:"semester"`
	Nominal       float64 `json:"nominal"`
	JatuhTempo    string  `json:"jatuh_tempo"`
	Catatan       string  `json:"catatan"`
}

type CreatePembayaranRequest struct {
	TagihanID         int64   `json:"tagihan_id"`
	SiswaID           int64   `json:"siswa_id"`
	Jumlah            float64 `json:"jumlah"`
	Tanggal           string  `json:"tanggal"`
	Metode            string  `json:"metode"`
	Provider          string  `json:"provider"`
	BuktiBayar        string  `json:"bukti_bayar"`
	PaymentGatewayID  string  `json:"payment_gateway_id"`
	RekeningSekolahID int64   `json:"rekening_sekolah_id"`
	Catatan           string  `json:"catatan"`
}

func (s *Service) ListKategori(sekolahID int64) ([]Kategori, error) {
	return s.repo.ListKategori(sekolahID)
}

func (s *Service) GetKategoriByID(sekolahID, id int64) (*Kategori, error) {
	return s.repo.GetKategoriByID(sekolahID, id)
}

func (s *Service) CreateKategori(sekolahID int64, req CreateKategoriRequest) (*Kategori, error) {
	errs := validator.Collect(
		validator.Required("nama", req.Nama),
	)
	if len(errs) > 0 {
		return nil, errs
	}

	k := &Kategori{
		SekolahID: sekolahID,
		Nama:      req.Nama,
		Deskripsi: req.Deskripsi,
	}

	id, err := s.repo.CreateKategori(k)
	if err != nil {
		return nil, fmt.Errorf("create kategori: %w", err)
	}

	return s.repo.GetKategoriByID(sekolahID, id)
}

func (s *Service) UpdateKategori(sekolahID, id int64, req UpdateKategoriRequest) (*Kategori, error) {
	errs := validator.Collect(
		validator.Required("nama", req.Nama),
	)
	if len(errs) > 0 {
		return nil, errs
	}

	k := &Kategori{
		Nama:      req.Nama,
		Deskripsi: req.Deskripsi,
	}

	if err := s.repo.UpdateKategori(sekolahID, id, k); err != nil {
		return nil, fmt.Errorf("update kategori: %w", err)
	}

	return s.repo.GetKategoriByID(sekolahID, id)
}

func (s *Service) DeleteKategori(sekolahID, id int64) error {
	_, err := s.repo.GetKategoriByID(sekolahID, id)
	if err != nil {
		return fmt.Errorf("kategori tidak ditemukan")
	}
	return s.repo.DeleteKategori(sekolahID, id)
}

func (s *Service) ListTagihan(sekolahID int64, params TagihanListParams) ([]Tagihan, int, error) {
	return s.repo.ListTagihan(sekolahID, params)
}

func (s *Service) GetTagihanByID(sekolahID, id int64) (*Tagihan, error) {
	return s.repo.GetTagihanByID(sekolahID, id)
}

func (s *Service) CreateTagihan(sekolahID int64, req CreateTagihanRequest) (*Tagihan, error) {
	var errs validator.ValidationErrors
	if req.SiswaID == 0 {
		errs = append(errs, validator.ValidationError{Field: "siswa_id", Message: "wajib diisi"})
	}
	if req.KategoriID == 0 {
		errs = append(errs, validator.ValidationError{Field: "kategori_id", Message: "wajib diisi"})
	}
	if req.TahunAjaranID == 0 {
		errs = append(errs, validator.ValidationError{Field: "tahun_ajaran_id", Message: "wajib diisi"})
	}
	if req.Nominal <= 0 {
		errs = append(errs, validator.ValidationError{Field: "nominal", Message: "harus lebih dari 0"})
	}
	if len(errs) > 0 {
		return nil, errs
	}

	t := &Tagihan{
		SekolahID:     sekolahID,
		SiswaID:       req.SiswaID,
		KategoriID:    req.KategoriID,
		TahunAjaranID: req.TahunAjaranID,
		Semester:      req.Semester,
		Nominal:       req.Nominal,
		JatuhTempo:    req.JatuhTempo,
		Catatan:       req.Catatan,
	}

	id, err := s.repo.CreateTagihan(t)
	if err != nil {
		return nil, fmt.Errorf("create tagihan: %w", err)
	}

	return s.repo.GetTagihanByID(sekolahID, id)
}

func (s *Service) BulkCreateTagihan(sekolahID int64, req BulkCreateTagihanRequest) error {
	if len(req.SiswaIDs) == 0 {
		return validator.ValidationErrors{{Field: "siswa_ids", Message: "wajib diisi"}}
	}
	if req.KategoriID == 0 {
		return validator.ValidationErrors{{Field: "kategori_id", Message: "wajib diisi"}}
	}
	if req.TahunAjaranID == 0 {
		return validator.ValidationErrors{{Field: "tahun_ajaran_id", Message: "wajib diisi"}}
	}
	if req.Nominal <= 0 {
		return validator.ValidationErrors{{Field: "nominal", Message: "harus lebih dari 0"}}
	}

	var items []*Tagihan
	for _, siswaID := range req.SiswaIDs {
		items = append(items, &Tagihan{
			SekolahID:     sekolahID,
			SiswaID:       siswaID,
			KategoriID:    req.KategoriID,
			TahunAjaranID: req.TahunAjaranID,
			Semester:      req.Semester,
			Nominal:       req.Nominal,
			JatuhTempo:    req.JatuhTempo,
			Catatan:       req.Catatan,
		})
	}

	return s.repo.BulkCreateTagihan(items)
}

func (s *Service) UpdateTagihan(sekolahID, id int64, req UpdateTagihanRequest) (*Tagihan, error) {
	if req.KategoriID == 0 {
		return nil, validator.ValidationErrors{{Field: "kategori_id", Message: "wajib diisi"}}
	}
	if req.TahunAjaranID == 0 {
		return nil, validator.ValidationErrors{{Field: "tahun_ajaran_id", Message: "wajib diisi"}}
	}
	if req.Nominal <= 0 {
		return nil, validator.ValidationErrors{{Field: "nominal", Message: "harus lebih dari 0"}}
	}

	t := &Tagihan{
		KategoriID:    req.KategoriID,
		TahunAjaranID: req.TahunAjaranID,
		Semester:      req.Semester,
		Nominal:       req.Nominal,
		JatuhTempo:    req.JatuhTempo,
		Catatan:       req.Catatan,
	}

	if err := s.repo.UpdateTagihan(sekolahID, id, t); err != nil {
		return nil, fmt.Errorf("update tagihan: %w", err)
	}

	return s.repo.GetTagihanByID(sekolahID, id)
}

func (s *Service) DeleteTagihan(sekolahID, id int64) error {
	_, err := s.repo.GetTagihanByID(sekolahID, id)
	if err != nil {
		return fmt.Errorf("tagihan tidak ditemukan")
	}
	return s.repo.DeleteTagihan(sekolahID, id)
}

func (s *Service) ListPembayaran(sekolahID int64, params PembayaranListParams) ([]Pembayaran, int, error) {
	return s.repo.ListPembayaran(sekolahID, params)
}

func (s *Service) GetPembayaranByID(sekolahID, id int64) (*Pembayaran, error) {
	return s.repo.GetPembayaranByID(sekolahID, id)
}

func (s *Service) CreatePembayaran(sekolahID int64, req CreatePembayaranRequest) (*Pembayaran, error) {
	allowedMetode := []string{"transfer", "cash", "midtrans", "xendit"}
	errs := validator.Collect(
		validator.Required("tanggal", req.Tanggal),
		validator.Required("metode", req.Metode),
		validator.InList("metode", req.Metode, allowedMetode),
	)
	if req.TagihanID == 0 {
		errs = append(errs, validator.ValidationError{Field: "tagihan_id", Message: "wajib diisi"})
	}
	if req.SiswaID == 0 {
		errs = append(errs, validator.ValidationError{Field: "siswa_id", Message: "wajib diisi"})
	}
	if req.Jumlah <= 0 {
		errs = append(errs, validator.ValidationError{Field: "jumlah", Message: "harus lebih dari 0"})
	}
	if req.Metode == "transfer" && req.RekeningSekolahID == 0 {
		errs = append(errs, validator.ValidationError{Field: "rekening_sekolah_id", Message: "wajib diisi untuk metode transfer"})
	}
	if len(errs) > 0 {
		return nil, errs
	}

	p := &Pembayaran{
		TagihanID:         req.TagihanID,
		SiswaID:           req.SiswaID,
		Jumlah:            req.Jumlah,
		Tanggal:           req.Tanggal,
		Metode:            req.Metode,
		Provider:          req.Provider,
		BuktiBayar:        req.BuktiBayar,
		PaymentGatewayID:  req.PaymentGatewayID,
		RekeningSekolahID: req.RekeningSekolahID,
		Catatan:           req.Catatan,
	}

	id, err := s.repo.CreatePembayaran(p)
	if err != nil {
		return nil, fmt.Errorf("create pembayaran: %w", err)
	}

	return s.repo.GetPembayaranByID(sekolahID, id)
}

func (s *Service) VerifyPembayaran(sekolahID, id, verifiedBy int64) error {
	p, err := s.repo.GetPembayaranByID(sekolahID, id)
	if err != nil {
		return fmt.Errorf("pembayaran tidak ditemukan")
	}

	tagihan, err := s.repo.GetTagihanByID(sekolahID, p.TagihanID)
	if err != nil {
		return fmt.Errorf("tagihan tidak ditemukan")
	}

	verifiedSum, err := s.repo.GetVerifiedSum(p.TagihanID)
	if err != nil {
		return fmt.Errorf("gagal cek total pembayaran: %w", err)
	}

	if verifiedSum+p.Jumlah > tagihan.Nominal {
		return validator.ValidationErrors{{Field: "jumlah", Message: "total pembayaran melebihi nominal tagihan"}}
	}

	if err := s.repo.VerifyPembayaran(sekolahID, id, verifiedBy); err != nil {
		return fmt.Errorf("verify pembayaran: %w", err)
	}

	newSum := verifiedSum + p.Jumlah
	var status string
	if newSum >= tagihan.Nominal {
		status = "lunas"
	} else if newSum > 0 {
		status = "sebagian"
	} else {
		status = "belum_bayar"
	}

	return s.repo.UpdateTagihanStatus(sekolahID, p.TagihanID, status)
}

func (s *Service) RejectPembayaran(sekolahID, id int64) error {
	_, err := s.repo.GetPembayaranByID(sekolahID, id)
	if err != nil {
		return fmt.Errorf("pembayaran tidak ditemukan")
	}
	return s.repo.RejectPembayaran(sekolahID, id)
}
