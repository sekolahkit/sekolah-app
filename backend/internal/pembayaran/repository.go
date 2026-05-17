package pembayaran

import (
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type Kategori struct {
	ID        int64  `json:"id"`
	SekolahID int64  `json:"sekolah_id"`
	Nama      string `json:"nama"`
	Deskripsi string `json:"deskripsi"`
	CreatedAt string `json:"created_at"`
}

type Tagihan struct {
	ID            int64   `json:"id"`
	SekolahID     int64   `json:"sekolah_id"`
	SiswaID       int64   `json:"siswa_id"`
	KategoriID    int64   `json:"kategori_id"`
	TahunAjaranID int64   `json:"tahun_ajaran_id"`
	Semester      string  `json:"semester"`
	Nominal       float64 `json:"nominal"`
	JatuhTempo    string  `json:"jatuh_tempo"`
	Status        string  `json:"status"`
	Catatan       string  `json:"catatan"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type Pembayaran struct {
	ID                int64   `json:"id"`
	TagihanID         int64   `json:"tagihan_id"`
	SiswaID           int64   `json:"siswa_id"`
	Jumlah            float64 `json:"jumlah"`
	Tanggal           string  `json:"tanggal"`
	Metode            string  `json:"metode"`
	Provider          string  `json:"provider"`
	BuktiBayar        string  `json:"bukti_bayar"`
	PaymentGatewayID  string  `json:"payment_gateway_id"`
	RekeningSekolahID int64   `json:"rekening_sekolah_id"`
	Status            string  `json:"status"`
	Catatan           string  `json:"catatan"`
	VerifiedBy        int64   `json:"verified_by"`
	VerifiedAt        string  `json:"verified_at"`
	CreatedAt         string  `json:"created_at"`
}

type TagihanListParams struct {
	Page          int
	Limit         int
	Sort          string
	Order         string
	SiswaID       int64
	KategoriID    int64
	TahunAjaranID int64
	Status        string
}

type PembayaranListParams struct {
	Page      int
	Limit     int
	Sort      string
	Order     string
	TagihanID int64
	Status    string
}

func (r *Repository) ListKategori(sekolahID int64) ([]Kategori, error) {
	rows, err := sq.Select("id", "sekolah_id", "nama", "COALESCE(deskripsi,'')", "created_at").
		From("kategori_pembayaran").
		Where(sq.Eq{"sekolah_id": sekolahID}).
		OrderBy("nama ASC").
		RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Kategori
	for rows.Next() {
		var k Kategori
		if err := rows.Scan(&k.ID, &k.SekolahID, &k.Nama, &k.Deskripsi, &k.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, k)
	}
	return list, nil
}

func (r *Repository) GetKategoriByID(sekolahID, id int64) (*Kategori, error) {
	var k Kategori
	err := sq.Select("id", "sekolah_id", "nama", "COALESCE(deskripsi,'')", "created_at").
		From("kategori_pembayaran").
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).QueryRow().
		Scan(&k.ID, &k.SekolahID, &k.Nama, &k.Deskripsi, &k.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *Repository) CreateKategori(k *Kategori) (int64, error) {
	result, err := sq.Insert("kategori_pembayaran").
		Columns("sekolah_id", "nama", "deskripsi").
		Values(k.SekolahID, k.Nama, k.Deskripsi).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) UpdateKategori(sekolahID, id int64, k *Kategori) error {
	_, err := sq.Update("kategori_pembayaran").
		Set("nama", k.Nama).
		Set("deskripsi", k.Deskripsi).
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) DeleteKategori(sekolahID, id int64) error {
	_, err := sq.Delete("kategori_pembayaran").
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) ListTagihan(sekolahID int64, params TagihanListParams) ([]Tagihan, int, error) {
	query := sq.Select("id", "sekolah_id", "siswa_id", "kategori_id", "tahun_ajaran_id",
		"COALESCE(semester,'')", "nominal", "COALESCE(jatuh_tempo,'')", "status",
		"COALESCE(catatan,'')", "created_at", "updated_at").
		From("tagihan").
		Where(sq.Eq{"sekolah_id": sekolahID})

	countQuery := sq.Select("COUNT(*)").From("tagihan").Where(sq.Eq{"sekolah_id": sekolahID})

	if params.SiswaID > 0 {
		query = query.Where(sq.Eq{"siswa_id": params.SiswaID})
		countQuery = countQuery.Where(sq.Eq{"siswa_id": params.SiswaID})
	}
	if params.KategoriID > 0 {
		query = query.Where(sq.Eq{"kategori_id": params.KategoriID})
		countQuery = countQuery.Where(sq.Eq{"kategori_id": params.KategoriID})
	}
	if params.TahunAjaranID > 0 {
		query = query.Where(sq.Eq{"tahun_ajaran_id": params.TahunAjaranID})
		countQuery = countQuery.Where(sq.Eq{"tahun_ajaran_id": params.TahunAjaranID})
	}
	if params.Status != "" {
		query = query.Where(sq.Eq{"status": params.Status})
		countQuery = countQuery.Where(sq.Eq{"status": params.Status})
	}

	var total int
	if err := countQuery.RunWith(r.db).QueryRow().Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Limit
	query = query.OrderBy(params.Sort + " " + params.Order).Limit(uint64(params.Limit)).Offset(uint64(offset))

	rows, err := query.RunWith(r.db).Query()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []Tagihan
	for rows.Next() {
		var t Tagihan
		if err := rows.Scan(&t.ID, &t.SekolahID, &t.SiswaID, &t.KategoriID, &t.TahunAjaranID,
			&t.Semester, &t.Nominal, &t.JatuhTempo, &t.Status, &t.Catatan, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, t)
	}
	return list, total, nil
}

func (r *Repository) GetTagihanByID(sekolahID, id int64) (*Tagihan, error) {
	var t Tagihan
	err := sq.Select("id", "sekolah_id", "siswa_id", "kategori_id", "tahun_ajaran_id",
		"COALESCE(semester,'')", "nominal", "COALESCE(jatuh_tempo,'')", "status",
		"COALESCE(catatan,'')", "created_at", "updated_at").
		From("tagihan").
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).QueryRow().
		Scan(&t.ID, &t.SekolahID, &t.SiswaID, &t.KategoriID, &t.TahunAjaranID,
			&t.Semester, &t.Nominal, &t.JatuhTempo, &t.Status, &t.Catatan, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) CreateTagihan(t *Tagihan) (int64, error) {
	result, err := sq.Insert("tagihan").
		Columns("sekolah_id", "siswa_id", "kategori_id", "tahun_ajaran_id", "semester", "nominal", "jatuh_tempo", "catatan").
		Values(t.SekolahID, t.SiswaID, t.KategoriID, t.TahunAjaranID, t.Semester, t.Nominal, t.JatuhTempo, t.Catatan).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) BulkCreateTagihan(items []*Tagihan) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, t := range items {
		_, err := sq.Insert("tagihan").
			Columns("sekolah_id", "siswa_id", "kategori_id", "tahun_ajaran_id", "semester", "nominal", "jatuh_tempo", "catatan").
			Values(t.SekolahID, t.SiswaID, t.KategoriID, t.TahunAjaranID, t.Semester, t.Nominal, t.JatuhTempo, t.Catatan).
			RunWith(tx).Exec()
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) UpdateTagihan(sekolahID, id int64, t *Tagihan) error {
	_, err := sq.Update("tagihan").
		Set("kategori_id", t.KategoriID).
		Set("tahun_ajaran_id", t.TahunAjaranID).
		Set("semester", t.Semester).
		Set("nominal", t.Nominal).
		Set("jatuh_tempo", t.JatuhTempo).
		Set("catatan", t.Catatan).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) DeleteTagihan(sekolahID, id int64) error {
	_, err := sq.Delete("tagihan").
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) UpdateTagihanStatus(sekolahID, id int64, status string) error {
	_, err := sq.Update("tagihan").
		Set("status", status).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) ListPembayaran(sekolahID int64, params PembayaranListParams) ([]Pembayaran, int, error) {
	query := sq.Select("p.id", "p.tagihan_id", "p.siswa_id", "p.jumlah", "p.tanggal",
		"p.metode", "COALESCE(p.provider,'')", "COALESCE(p.bukti_bayar,'')",
		"COALESCE(p.payment_gateway_id,'')", "COALESCE(p.rekening_sekolah_id,0)",
		"p.status", "COALESCE(p.catatan,'')", "COALESCE(p.verified_by,0)",
		"COALESCE(p.verified_at,'')", "p.created_at").
		From("pembayaran p").
		Join("tagihan t ON t.id = p.tagihan_id").
		Where(sq.Eq{"t.sekolah_id": sekolahID})

	countQuery := sq.Select("COUNT(*)").
		From("pembayaran p").
		Join("tagihan t ON t.id = p.tagihan_id").
		Where(sq.Eq{"t.sekolah_id": sekolahID})

	if params.TagihanID > 0 {
		query = query.Where(sq.Eq{"p.tagihan_id": params.TagihanID})
		countQuery = countQuery.Where(sq.Eq{"p.tagihan_id": params.TagihanID})
	}
	if params.Status != "" {
		query = query.Where(sq.Eq{"p.status": params.Status})
		countQuery = countQuery.Where(sq.Eq{"p.status": params.Status})
	}

	var total int
	if err := countQuery.RunWith(r.db).QueryRow().Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Limit
	query = query.OrderBy("p." + params.Sort + " " + params.Order).Limit(uint64(params.Limit)).Offset(uint64(offset))

	rows, err := query.RunWith(r.db).Query()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []Pembayaran
	for rows.Next() {
		var p Pembayaran
		if err := rows.Scan(&p.ID, &p.TagihanID, &p.SiswaID, &p.Jumlah, &p.Tanggal,
			&p.Metode, &p.Provider, &p.BuktiBayar, &p.PaymentGatewayID, &p.RekeningSekolahID,
			&p.Status, &p.Catatan, &p.VerifiedBy, &p.VerifiedAt, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, p)
	}
	return list, total, nil
}

func (r *Repository) GetPembayaranByID(sekolahID, id int64) (*Pembayaran, error) {
	var p Pembayaran
	err := sq.Select("p.id", "p.tagihan_id", "p.siswa_id", "p.jumlah", "p.tanggal",
		"p.metode", "COALESCE(p.provider,'')", "COALESCE(p.bukti_bayar,'')",
		"COALESCE(p.payment_gateway_id,'')", "COALESCE(p.rekening_sekolah_id,0)",
		"p.status", "COALESCE(p.catatan,'')", "COALESCE(p.verified_by,0)",
		"COALESCE(p.verified_at,'')", "p.created_at").
		From("pembayaran p").
		Join("tagihan t ON t.id = p.tagihan_id").
		Where(sq.Eq{"p.id": id, "t.sekolah_id": sekolahID}).
		RunWith(r.db).QueryRow().
		Scan(&p.ID, &p.TagihanID, &p.SiswaID, &p.Jumlah, &p.Tanggal,
			&p.Metode, &p.Provider, &p.BuktiBayar, &p.PaymentGatewayID, &p.RekeningSekolahID,
			&p.Status, &p.Catatan, &p.VerifiedBy, &p.VerifiedAt, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) CreatePembayaran(p *Pembayaran) (int64, error) {
	var rekeningID interface{}
	if p.RekeningSekolahID > 0 {
		rekeningID = p.RekeningSekolahID
	}
	var provider interface{}
	if p.Provider != "" {
		provider = p.Provider
	}
	var gatewayID interface{}
	if p.PaymentGatewayID != "" {
		gatewayID = p.PaymentGatewayID
	}
	var buktiBayar interface{}
	if p.BuktiBayar != "" {
		buktiBayar = p.BuktiBayar
	}

	result, err := sq.Insert("pembayaran").
		Columns("tagihan_id", "siswa_id", "jumlah", "tanggal", "metode", "provider",
			"bukti_bayar", "payment_gateway_id", "rekening_sekolah_id", "catatan").
		Values(p.TagihanID, p.SiswaID, p.Jumlah, p.Tanggal, p.Metode, provider,
			buktiBayar, gatewayID, rekeningID, p.Catatan).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) VerifyPembayaran(sekolahID, id, verifiedBy int64) error {
	_, err := sq.Update("pembayaran").
		Set("status", "verified").
		Set("verified_by", verifiedBy).
		Set("verified_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": id}).
		Where("tagihan_id IN (SELECT id FROM tagihan WHERE sekolah_id = ?)", sekolahID).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) RejectPembayaran(sekolahID, id int64) error {
	_, err := sq.Update("pembayaran").
		Set("status", "rejected").
		Where(sq.Eq{"id": id}).
		Where("tagihan_id IN (SELECT id FROM tagihan WHERE sekolah_id = ?)", sekolahID).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) GetVerifiedSum(tagihanID int64) (float64, error) {
	var sum sql.NullFloat64
	err := sq.Select("COALESCE(SUM(jumlah),0)").
		From("pembayaran").
		Where(sq.Eq{"tagihan_id": tagihanID, "status": "verified"}).
		RunWith(r.db).QueryRow().Scan(&sum)
	if err != nil {
		return 0, err
	}
	return sum.Float64, nil
}

func (r *Repository) VerifyPembayaranTx(sekolahID, id, verifiedBy int64) (string, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var p Pembayaran
	err = sq.Select("p.id", "p.tagihan_id", "p.siswa_id", "p.jumlah", "p.tanggal",
		"p.metode", "COALESCE(p.provider,'')", "COALESCE(p.bukti_bayar,'')",
		"COALESCE(p.payment_gateway_id,'')", "COALESCE(p.rekening_sekolah_id,0)",
		"p.status", "COALESCE(p.catatan,'')", "COALESCE(p.verified_by,0)",
		"COALESCE(p.verified_at,'')", "p.created_at").
		From("pembayaran p").
		Join("tagihan t ON t.id = p.tagihan_id").
		Where(sq.Eq{"p.id": id, "t.sekolah_id": sekolahID}).
		RunWith(tx).QueryRow().
		Scan(&p.ID, &p.TagihanID, &p.SiswaID, &p.Jumlah, &p.Tanggal,
			&p.Metode, &p.Provider, &p.BuktiBayar, &p.PaymentGatewayID, &p.RekeningSekolahID,
			&p.Status, &p.Catatan, &p.VerifiedBy, &p.VerifiedAt, &p.CreatedAt)
	if err != nil {
		return "", fmt.Errorf("pembayaran tidak ditemukan")
	}

	if p.Status != "pending" {
		return "", fmt.Errorf("pembayaran sudah diproses")
	}

	var tagihan Tagihan
	err = sq.Select("id", "sekolah_id", "siswa_id", "kategori_id", "tahun_ajaran_id",
		"COALESCE(semester,'')", "nominal", "COALESCE(jatuh_tempo,'')", "status",
		"COALESCE(catatan,'')", "created_at", "updated_at").
		From("tagihan").
		Where(sq.Eq{"id": p.TagihanID, "sekolah_id": sekolahID}).
		RunWith(tx).QueryRow().
		Scan(&tagihan.ID, &tagihan.SekolahID, &tagihan.SiswaID, &tagihan.KategoriID, &tagihan.TahunAjaranID,
			&tagihan.Semester, &tagihan.Nominal, &tagihan.JatuhTempo, &tagihan.Status, &tagihan.Catatan, &tagihan.CreatedAt, &tagihan.UpdatedAt)
	if err != nil {
		return "", fmt.Errorf("tagihan tidak ditemukan")
	}

	var verifiedSum sql.NullFloat64
	err = sq.Select("COALESCE(SUM(jumlah),0)").
		From("pembayaran").
		Where(sq.Eq{"tagihan_id": p.TagihanID, "status": "verified"}).
		RunWith(tx).QueryRow().Scan(&verifiedSum)
	if err != nil {
		return "", err
	}

	if verifiedSum.Float64+p.Jumlah > tagihan.Nominal {
		return "", fmt.Errorf("total pembayaran melebihi nominal tagihan")
	}

	_, err = sq.Update("pembayaran").
		Set("status", "verified").
		Set("verified_by", verifiedBy).
		Set("verified_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": id}).
		RunWith(tx).Exec()
	if err != nil {
		return "", err
	}

	newSum := verifiedSum.Float64 + p.Jumlah
	var status string
	if newSum >= tagihan.Nominal {
		status = "lunas"
	} else if newSum > 0 {
		status = "sebagian"
	} else {
		status = "belum_bayar"
	}

	_, err = sq.Update("tagihan").
		Set("status", status).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": p.TagihanID, "sekolah_id": sekolahID}).
		RunWith(tx).Exec()
	if err != nil {
		return "", err
	}

	return status, tx.Commit()
}

func (r *Repository) SiswaExistsInSekolah(sekolahID, siswaID int64) bool {
	var count int
	err := sq.Select("COUNT(*)").From("siswa").
		Where(sq.Eq{"id": siswaID, "sekolah_id": sekolahID}).
		RunWith(r.db).QueryRow().Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

func (r *Repository) KategoriExistsInSekolah(sekolahID, kategoriID int64) bool {
	var count int
	err := sq.Select("COUNT(*)").From("kategori_pembayaran").
		Where(sq.Eq{"id": kategoriID, "sekolah_id": sekolahID}).
		RunWith(r.db).QueryRow().Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}
