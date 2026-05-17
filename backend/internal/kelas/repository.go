package kelas

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

type Kelas struct {
	ID            int64  `json:"id"`
	SekolahID     int64  `json:"sekolah_id"`
	TahunAjaranID int64  `json:"tahun_ajaran_id"`
	Nama          string `json:"nama"`
	Tingkat       string `json:"tingkat"`
	JurusanID     int64  `json:"jurusan_id"`
	WaliKelasID   *int64 `json:"wali_kelas_id"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type KelasSiswa struct {
	ID            int64 `json:"id"`
	KelasID       int64 `json:"kelas_id"`
	SiswaID       int64 `json:"siswa_id"`
	SekolahID     int64 `json:"sekolah_id"`
	TahunAjaranID int64 `json:"tahun_ajaran_id"`
}

type ListParams struct {
	Page          int
	Limit         int
	Search        string
	Sort          string
	Order         string
	TahunAjaranID int64
}

func (r *Repository) List(sekolahID int64, params ListParams) ([]Kelas, int, error) {
	query := sq.Select("id", "sekolah_id", "tahun_ajaran_id", "nama", "tingkat",
		"COALESCE(jurusan_id,0)", "COALESCE(wali_kelas_id,0)", "created_at", "updated_at").
		From("kelas").
		Where(sq.Eq{"sekolah_id": sekolahID})

	countQuery := sq.Select("COUNT(*)").From("kelas").Where(sq.Eq{"sekolah_id": sekolahID})

	if params.TahunAjaranID > 0 {
		query = query.Where(sq.Eq{"tahun_ajaran_id": params.TahunAjaranID})
		countQuery = countQuery.Where(sq.Eq{"tahun_ajaran_id": params.TahunAjaranID})
	}

	if params.Search != "" {
		like := "%" + params.Search + "%"
		query = query.Where(sq.Like{"nama": like})
		countQuery = countQuery.Where(sq.Like{"nama": like})
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

	var list []Kelas
	for rows.Next() {
		var k Kelas
		var waliID int64
		err := rows.Scan(&k.ID, &k.SekolahID, &k.TahunAjaranID, &k.Nama, &k.Tingkat, &k.JurusanID, &waliID, &k.CreatedAt, &k.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		if waliID != 0 {
			k.WaliKelasID = &waliID
		}
		list = append(list, k)
	}
	return list, total, nil
}

func (r *Repository) GetByID(sekolahID, id int64) (*Kelas, error) {
	var k Kelas
	var waliID int64
	err := sq.Select("id", "sekolah_id", "tahun_ajaran_id", "nama", "tingkat",
		"COALESCE(jurusan_id,0)", "COALESCE(wali_kelas_id,0)", "created_at", "updated_at").
		From("kelas").
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).QueryRow().
		Scan(&k.ID, &k.SekolahID, &k.TahunAjaranID, &k.Nama, &k.Tingkat, &k.JurusanID, &waliID, &k.CreatedAt, &k.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if waliID != 0 {
		k.WaliKelasID = &waliID
	}
	return &k, nil
}

func (r *Repository) Create(k *Kelas) (int64, error) {
	result, err := sq.Insert("kelas").
		Columns("sekolah_id", "tahun_ajaran_id", "nama", "tingkat", "jurusan_id", "wali_kelas_id").
		Values(k.SekolahID, k.TahunAjaranID, k.Nama, k.Tingkat, nullInt64(k.JurusanID), k.WaliKelasID).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(sekolahID, id int64, k *Kelas) error {
	_, err := sq.Update("kelas").
		Set("tahun_ajaran_id", k.TahunAjaranID).
		Set("nama", k.Nama).
		Set("tingkat", k.Tingkat).
		Set("jurusan_id", nullInt64(k.JurusanID)).
		Set("wali_kelas_id", k.WaliKelasID).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).Exec()
	return err
}

func nullInt64(v int64) interface{} {
	if v == 0 {
		return nil
	}
	return v
}

func (r *Repository) Delete(sekolahID, id int64) error {
	_, err := sq.Delete("kelas").
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) AddSiswa(sekolahID, kelasID, siswaID, tahunAjaranID int64) error {
	_, err := sq.Insert("kelas_siswa").
		Columns("sekolah_id", "kelas_id", "siswa_id", "tahun_ajaran_id").
		Values(sekolahID, kelasID, siswaID, tahunAjaranID).
		Suffix("ON CONFLICT(sekolah_id, siswa_id, kelas_id, tahun_ajaran_id) DO NOTHING").
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) RemoveSiswa(sekolahID, kelasID, siswaID int64) error {
	_, err := sq.Delete("kelas_siswa").
		Where(sq.Eq{"sekolah_id": sekolahID, "kelas_id": kelasID, "siswa_id": siswaID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) ListSiswa(sekolahID, kelasID int64) ([]int64, error) {
	rows, err := sq.Select("siswa_id").From("kelas_siswa").
		Where(sq.Eq{"sekolah_id": sekolahID, "kelas_id": kelasID}).
		RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
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

func (r *Repository) SiswaInKelas(sekolahID, kelasID, siswaID int64) (bool, error) {
	var count int
	err := sq.Select("COUNT(*)").From("kelas_siswa").
		Where(sq.Eq{"sekolah_id": sekolahID, "kelas_id": kelasID, "siswa_id": siswaID}).
		RunWith(r.db).QueryRow().Scan(&count)
	return count > 0, err
}
