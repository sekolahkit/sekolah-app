package sekolah

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

type Sekolah struct {
	ID                 int64  `json:"id"`
	Nama               string `json:"nama"`
	Kode               string `json:"kode"`
	Alamat             string `json:"alamat"`
	Telepon            string `json:"telepon"`
	Email              string `json:"email"`
	Logo               string `json:"logo"`
	Website            string `json:"website"`
	NamaKepalaSekolah  string `json:"nama_kepala_sekolah"`
	NIPKepalaSekolah   string `json:"nip_kepala_sekolah"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

func (r *Repository) GetByID(id int64) (*Sekolah, error) {
	var s Sekolah
	err := sq.Select("id", "nama", "kode",
		"COALESCE(alamat,'')", "COALESCE(telepon,'')", "COALESCE(email,'')",
		"COALESCE(logo,'')", "COALESCE(website,'')",
		"COALESCE(nama_kepala_sekolah,'')", "COALESCE(nip_kepala_sekolah,'')",
		"created_at", "updated_at").
		From("sekolah").
		Where(sq.Eq{"id": id}).
		RunWith(r.db).QueryRow().
		Scan(&s.ID, &s.Nama, &s.Kode, &s.Alamat, &s.Telepon, &s.Email,
			&s.Logo, &s.Website, &s.NamaKepalaSekolah, &s.NIPKepalaSekolah,
			&s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) Update(id int64, s *Sekolah) error {
	_, err := sq.Update("sekolah").
		Set("nama", s.Nama).
		Set("alamat", s.Alamat).
		Set("telepon", s.Telepon).
		Set("email", s.Email).
		Set("logo", s.Logo).
		Set("website", s.Website).
		Set("nama_kepala_sekolah", s.NamaKepalaSekolah).
		Set("nip_kepala_sekolah", s.NIPKepalaSekolah).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": id}).
		RunWith(r.db).Exec()
	return err
}
