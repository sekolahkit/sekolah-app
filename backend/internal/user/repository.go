package user

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type User struct {
	ID        int64     `json:"id"`
	SekolahID int64     `json:"sekolah_id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Nama      string    `json:"nama"`
	Role      string    `json:"role"`
	GoogleID  *string   `json:"google_id,omitempty"`
	Foto      *string   `json:"foto,omitempty"`
	NoHP      *string   `json:"no_hp,omitempty"`
	Aktif     bool      `json:"aktif"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListParams struct {
	Page   int
	Limit  int
	Search string
	Role   string
	Aktif  *bool
}

func (r *Repository) List(sekolahID int64, params ListParams) ([]User, int, error) {
	base := sq.Select().From("pengguna").Where(sq.Eq{"sekolah_id": sekolahID})

	if params.Search != "" {
		like := "%" + params.Search + "%"
		base = base.Where(sq.Or{
			sq.Like{"nama": like},
			sq.Like{"email": like},
		})
	}
	if params.Role != "" {
		base = base.Where(sq.Eq{"role": params.Role})
	}
	if params.Aktif != nil {
		base = base.Where(sq.Eq{"aktif": *params.Aktif})
	}

	var total int
	countQuery := base.Column("COUNT(*)")
	err := countQuery.RunWith(r.db).QueryRow().Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Limit
	dataQuery := base.Columns("id", "sekolah_id", "email", "nama", "role", "google_id", "foto", "no_hp", "aktif", "created_at", "updated_at").
		OrderBy("nama ASC").
		Limit(uint64(params.Limit)).
		Offset(uint64(offset))

	rows, err := dataQuery.RunWith(r.db).Query()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.SekolahID, &u.Email, &u.Nama, &u.Role, &u.GoogleID, &u.Foto, &u.NoHP, &u.Aktif, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	return users, total, nil
}

func (r *Repository) GetByID(sekolahID, id int64) (*User, error) {
	u := &User{}
	err := sq.Select("id", "sekolah_id", "email", "nama", "role", "google_id", "foto", "no_hp", "aktif", "created_at", "updated_at").
		From("pengguna").
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).QueryRow().
		Scan(&u.ID, &u.SekolahID, &u.Email, &u.Nama, &u.Role, &u.GoogleID, &u.Foto, &u.NoHP, &u.Aktif, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) Create(u *User) (int64, error) {
	result, err := sq.Insert("pengguna").
		Columns("sekolah_id", "email", "password", "nama", "role", "no_hp", "aktif").
		Values(u.SekolahID, u.Email, u.Password, u.Nama, u.Role, u.NoHP, u.Aktif).
		RunWith(r.db).Exec()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(sekolahID, id int64, nama, email, role string, noHP *string, aktif bool) error {
	_, err := sq.Update("pengguna").
		Set("nama", nama).
		Set("email", email).
		Set("role", role).
		Set("no_hp", noHP).
		Set("aktif", aktif).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) UpdatePassword(sekolahID, id int64, hashedPassword string) error {
	_, err := sq.Update("pengguna").
		Set("password", hashedPassword).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": id, "sekolah_id": sekolahID}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) CountActiveAdmins(sekolahID int64) (int, error) {
	var count int
	err := sq.Select("COUNT(*)").From("pengguna").
		Where(sq.Eq{"sekolah_id": sekolahID, "role": "admin", "aktif": true}).
		RunWith(r.db).QueryRow().Scan(&count)
	return count, err
}

func (r *Repository) EmailExists(sekolahID int64, email string, excludeID *int64) (bool, error) {
	builder := sq.Select("COUNT(*)").From("pengguna").
		Where(sq.Eq{"sekolah_id": sekolahID, "email": email})
	if excludeID != nil {
		builder = builder.Where(sq.NotEq{"id": *excludeID})
	}
	var count int
	err := builder.RunWith(r.db).QueryRow().Scan(&count)
	return count > 0, err
}

func (r *Repository) GetPasswordByID(id int64) (string, error) {
	var hash string
	err := sq.Select("password").From("pengguna").
		Where(sq.Eq{"id": id}).
		RunWith(r.db).QueryRow().Scan(&hash)
	return hash, err
}
