package auth

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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
	ID        int64  `json:"id"`
	SekolahID int64  `json:"sekolah_id"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	Nama      string `json:"nama"`
	Role      string `json:"role"`
	Aktif     bool   `json:"aktif"`
}

type RefreshToken struct {
	ID         int64
	PenggunaID int64
	TokenHash  string
	DeviceInfo string
	ExpiresAt  time.Time
	RevokedAt  sql.NullTime
}

func (r *Repository) FindSekolahByKode(kode string) (int64, error) {
	var id int64
	err := sq.Select("id").From("sekolah").
		Where(sq.Eq{"kode": kode}).
		RunWith(r.db).QueryRow().Scan(&id)
	return id, err
}

func (r *Repository) FindUserByEmail(sekolahID int64, email string) (*User, error) {
	user := &User{}
	err := sq.Select("id", "sekolah_id", "email", "password", "nama", "role", "aktif").
		From("pengguna").
		Where(sq.Eq{"sekolah_id": sekolahID, "email": email}).
		RunWith(r.db).QueryRow().
		Scan(&user.ID, &user.SekolahID, &user.Email, &user.Password, &user.Nama, &user.Role, &user.Aktif)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) SaveRefreshToken(penggunaID int64, tokenHash, deviceInfo string, expiresAt time.Time) error {
	_, err := sq.Insert("refresh_token").
		Columns("pengguna_id", "token_hash", "device_info", "expires_at").
		Values(penggunaID, tokenHash, deviceInfo, expiresAt).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) FindRefreshToken(tokenHash string) (*RefreshToken, error) {
	rt := &RefreshToken{}
	err := sq.Select("id", "pengguna_id", "token_hash", "device_info", "expires_at", "revoked_at").
		From("refresh_token").
		Where(sq.Eq{"token_hash": tokenHash}).
		RunWith(r.db).QueryRow().
		Scan(&rt.ID, &rt.PenggunaID, &rt.TokenHash, &rt.DeviceInfo, &rt.ExpiresAt, &rt.RevokedAt)
	if err != nil {
		return nil, err
	}
	return rt, nil
}

func (r *Repository) RevokeRefreshToken(id int64) error {
	_, err := sq.Update("refresh_token").
		Set("revoked_at", time.Now()).
		Where(sq.Eq{"id": id}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) RevokeAllUserTokens(penggunaID int64) error {
	_, err := sq.Update("refresh_token").
		Set("revoked_at", time.Now()).
		Where(sq.Eq{"pengguna_id": penggunaID}).
		Where(sq.Eq{"revoked_at": nil}).
		RunWith(r.db).Exec()
	return err
}

func (r *Repository) RecordLoginAttempt(sekolahID *int64, email, ipAddress string, success bool) error {
	builder := sq.Insert("login_attempt").
		Columns("sekolah_id", "email", "ip_address", "success")
	if sekolahID != nil {
		builder = builder.Values(*sekolahID, email, ipAddress, success)
	} else {
		builder = builder.Values(nil, email, ipAddress, success)
	}
	_, err := builder.RunWith(r.db).Exec()
	return err
}

func (r *Repository) CountRecentFailedAttempts(sekolahID *int64, email string, since time.Time) (int, error) {
	builder := sq.Select("COUNT(*)").From("login_attempt").
		Where(sq.Eq{"email": email, "success": false}).
		Where(sq.GtOrEq{"created_at": since})
	if sekolahID != nil {
		builder = builder.Where(sq.Eq{"sekolah_id": *sekolahID})
	} else {
		builder = builder.Where("sekolah_id IS NULL")
	}
	var count int
	err := builder.RunWith(r.db).QueryRow().Scan(&count)
	return count, err
}

func (r *Repository) FindUserByID(id int64) (*User, error) {
	user := &User{}
	err := sq.Select("id", "sekolah_id", "email", "password", "nama", "role", "aktif").
		From("pengguna").
		Where(sq.Eq{"id": id}).
		RunWith(r.db).QueryRow().
		Scan(&user.ID, &user.SekolahID, &user.Email, &user.Password, &user.Nama, &user.Role, &user.Aktif)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
