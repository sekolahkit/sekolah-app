package siswa

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

type Siswa struct {
	ID             int64  `json:"id"`
	SekolahID      int64  `json:"sekolah_id"`
	NIS            string `json:"nis"`
	Nama           string `json:"nama"`
	JenisKelamin   string `json:"jenis_kelamin"`
	TempatLahir    string `json:"tempat_lahir"`
	TanggalLahir   string `json:"tanggal_lahir"`
	Agama          string `json:"agama"`
	Alamat         string `json:"alamat"`
	NoHP           string `json:"no_hp"`
	Email          string `json:"email"`
	NamaOrtu       string `json:"nama_ortu"`
	NoHPOrtu       string `json:"no_hp_ortu"`
	EmailOrtu      string `json:"email_ortu"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type ListParams struct {
	Page   int
	Limit  int
	Search string
	Sort   string
	Order  string
}

func (r *Repository) List(sekolahID int64, params ListParams) ([]Siswa, int, error) {
	query := sq.Select("id", "sekolah_id", "nis", "nama", "jenis_kelamin",
		"COALESCE(tempat_lahir,'')", "COALESCE(tanggal_lahir,'')", "COALESCE(agama,'')",
		"COALESCE(alamat,'')", "COALESCE(no_hp,'')", "COALESCE(email,'')",
		"COALESCE(nama_ortu,'')", "COALESCE(no_hp_ortu,'')", "COALESCE(email_ortu,'')",
		"status", "created_at", "updated_at").
		From("siswa").
		Where(sq.Eq{"sekolah_id": sekolahID})

	countQuery := sq.Select("COUNT(*)").From("siswa").Where(sq.Eq{"sekolah_id": sekolahID})

	if params.Search != "" {
		like := "%" + params.Search + "%"
		cond := sq.Or{sq.Like{"nama": like}, sq.Like{"nis": like}}
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

	var list []Siswa
	for rows.Next() {
		var s Siswa
		err := rows.Scan(&s.ID, &s.SekolahID, &s.NIS, &s.Nama, &s.JenisKelamin,
			&s.TempatLahir, &s.TanggalLahir, &s.Agama, &s.Alamat, &s.NoHP, &s.Email,
			&s.NamaOrtu, &s.NoHPOrtu, &s.EmailOrtu, &s.Status, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, s)
	}
	return list, total, nil
}

func (r *Repository) GetByID(sekolahID, id int64) (*Siswa, error) {
	var s Siswa
	err := sq.Select("id", "sekolah_id", "nis", "nama", "jenis_kelamin",
		"COALESCE(tempat_lahir,'')", "COALESCE(tanggal_lahir,'')", "COALESCE(agama,'')",
		"COALESCE(alamat,'')", "COALESCE(no_hp,'')", "COALESCE(email,'')",
		"COALESCE(nama_ortu,'')", "COALESCE(no_hp_ortu,'')", "COALESCE(email_ortu,'')",
		"status", "created_at", "updated_at").
		From("siswa").
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).QueryRow().
		Scan(&s.ID, &s.SekolahID, &s.NIS, &s.Nama, &s.JenisKelamin,
			&s.TempatLahir, &s.TanggalLahir, &s.Agama, &s.Alamat, &s.NoHP, &s.Email,
			&s.NamaOrtu, &s.NoHPOrtu, &s.EmailOrtu, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) Create(s *Siswa) (int64, error) {
	result, err := sq.Insert("siswa").
		Columns("sekolah_id", "nis", "nama", "jenis_kelamin", "tempat_lahir", "tanggal_lahir",
			"agama", "alamat", "no_hp", "email", "nama_ortu", "no_hp_ortu", "email_ortu").
		Values(s.SekolahID, s.NIS, s.Nama, s.JenisKelamin, s.TempatLahir, s.TanggalLahir,
			s.Agama, s.Alamat, s.NoHP, s.Email, s.NamaOrtu, s.NoHPOrtu, s.EmailOrtu).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(sekolahID, id int64, s *Siswa) error {
	_, err := sq.Update("siswa").
		Set("nis", s.NIS).
		Set("nama", s.Nama).
		Set("jenis_kelamin", s.JenisKelamin).
		Set("tempat_lahir", s.TempatLahir).
		Set("tanggal_lahir", s.TanggalLahir).
		Set("agama", s.Agama).
		Set("alamat", s.Alamat).
		Set("no_hp", s.NoHP).
		Set("email", s.Email).
		Set("nama_ortu", s.NamaOrtu).
		Set("no_hp_ortu", s.NoHPOrtu).
		Set("email_ortu", s.EmailOrtu).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) Delete(sekolahID, id int64) error {
	_, err := sq.Delete("siswa").
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) NISExists(sekolahID int64, nis string, excludeID int64) (bool, error) {
	query := sq.Select("COUNT(*)").From("siswa").Where(sq.Eq{"sekolah_id": sekolahID, "nis": nis})
	if excludeID > 0 {
		query = query.Where(sq.NotEq{"id": excludeID})
	}
	var count int
	err := query.RunWith(r.db).QueryRow().Scan(&count)
	return count > 0, err
}
