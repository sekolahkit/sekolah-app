package rekening

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type Rekening struct {
	ID            int64  `json:"id"`
	SekolahID     int64  `json:"sekolah_id"`
	NamaBank      string `json:"nama_bank"`
	NomorRekening string `json:"nomor_rekening"`
	NamaPemilik   string `json:"nama_pemilik"`
	Cabang        string `json:"cabang"`
	Aktif         bool   `json:"aktif"`
	Urutan        int    `json:"urutan"`
	Catatan       string `json:"catatan"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type ListParams struct {
	Page   int
	Limit  int
	Search string
	Sort   string
	Order  string
}

func (r *Repository) List(sekolahID int64, params ListParams) ([]Rekening, int, error) {
	query := sq.Select("id", "sekolah_id", "nama_bank", "nomor_rekening", "nama_pemilik",
		"COALESCE(cabang,'')", "aktif", "urutan", "COALESCE(catatan,'')", "created_at", "updated_at").
		From("rekening_sekolah").
		Where(sq.Eq{"sekolah_id": sekolahID})

	countQuery := sq.Select("COUNT(*)").From("rekening_sekolah").Where(sq.Eq{"sekolah_id": sekolahID})

	if params.Search != "" {
		like := "%" + params.Search + "%"
		cond := sq.Or{sq.Like{"nama_bank": like}, sq.Like{"nomor_rekening": like}, sq.Like{"nama_pemilik": like}}
		query = query.Where(cond)
		countQuery = countQuery.Where(cond)
	}

	var total int
	err := countQuery.RunWith(r.db).QueryRow().Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Limit
	query = query.OrderBy(params.Sort + " " + params.Order).Limit(uint64(params.Limit)).Offset(uint64(offset))

	rows, err := query.RunWith(r.db).Query()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []Rekening
	for rows.Next() {
		var rek Rekening
		err := rows.Scan(&rek.ID, &rek.SekolahID, &rek.NamaBank, &rek.NomorRekening, &rek.NamaPemilik,
			&rek.Cabang, &rek.Aktif, &rek.Urutan, &rek.Catatan, &rek.CreatedAt, &rek.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, rek)
	}
	return list, total, nil
}

func (r *Repository) GetByID(sekolahID, id int64) (*Rekening, error) {
	var rek Rekening
	err := sq.Select("id", "sekolah_id", "nama_bank", "nomor_rekening", "nama_pemilik",
		"COALESCE(cabang,'')", "aktif", "urutan", "COALESCE(catatan,'')", "created_at", "updated_at").
		From("rekening_sekolah").
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).QueryRow().
		Scan(&rek.ID, &rek.SekolahID, &rek.NamaBank, &rek.NomorRekening, &rek.NamaPemilik,
			&rek.Cabang, &rek.Aktif, &rek.Urutan, &rek.Catatan, &rek.CreatedAt, &rek.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &rek, nil
}

func (r *Repository) ListAktif(sekolahID int64) ([]Rekening, error) {
	rows, err := sq.Select("id", "sekolah_id", "nama_bank", "nomor_rekening", "nama_pemilik",
		"COALESCE(cabang,'')", "aktif", "urutan", "COALESCE(catatan,'')", "created_at", "updated_at").
		From("rekening_sekolah").
		Where(sq.Eq{"sekolah_id": sekolahID, "aktif": true}).
		OrderBy("urutan ASC", "nama_bank ASC").
		RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Rekening
	for rows.Next() {
		var rek Rekening
		err := rows.Scan(&rek.ID, &rek.SekolahID, &rek.NamaBank, &rek.NomorRekening, &rek.NamaPemilik,
			&rek.Cabang, &rek.Aktif, &rek.Urutan, &rek.Catatan, &rek.CreatedAt, &rek.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, rek)
	}
	return list, nil
}

func (r *Repository) Create(rek *Rekening) (int64, error) {
	result, err := sq.Insert("rekening_sekolah").
		Columns("sekolah_id", "nama_bank", "nomor_rekening", "nama_pemilik", "cabang", "aktif", "urutan", "catatan").
		Values(rek.SekolahID, rek.NamaBank, rek.NomorRekening, rek.NamaPemilik, rek.Cabang, true, rek.Urutan, rek.Catatan).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(sekolahID, id int64, rek *Rekening) error {
	_, err := sq.Update("rekening_sekolah").
		Set("nama_bank", rek.NamaBank).
		Set("nomor_rekening", rek.NomorRekening).
		Set("nama_pemilik", rek.NamaPemilik).
		Set("cabang", rek.Cabang).
		Set("urutan", rek.Urutan).
		Set("catatan", rek.Catatan).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) SoftDelete(sekolahID, id int64) error {
	_, err := sq.Update("rekening_sekolah").
		Set("aktif", false).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).Exec()
	return err
}
