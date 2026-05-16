package setup

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

func (r *Repository) IsInitialized() (bool, error) {
	var count int
	err := sq.Select("COUNT(*)").From("sekolah").RunWith(r.db).QueryRow().Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) CreateSekolah(nama, kode, alamat, telepon, email string) (int64, error) {
	result, err := sq.Insert("sekolah").
		Columns("nama", "kode", "alamat", "telepon", "email").
		Values(nama, kode, alamat, telepon, email).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) CreateAdmin(sekolahID int64, email, passwordHash, nama string) (int64, error) {
	result, err := sq.Insert("pengguna").
		Columns("sekolah_id", "email", "password", "nama", "role").
		Values(sekolahID, email, passwordHash, nama, "admin").
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
