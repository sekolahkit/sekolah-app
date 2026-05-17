package user

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var validRoles = []string{"admin", "operator", "guru", "siswa", "orangtua"}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreateRequest struct {
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Nama     string  `json:"nama"`
	Role     string  `json:"role"`
	NoHP     *string `json:"no_hp"`
}

type UpdateRequest struct {
	Email string  `json:"email"`
	Nama  string  `json:"nama"`
	Role  string  `json:"role"`
	NoHP  *string `json:"no_hp"`
	Aktif bool    `json:"aktif"`
}

type ResetPasswordRequest struct {
	Password string `json:"password"`
}

func (s *Service) List(sekolahID int64, params ListParams) ([]User, int, error) {
	return s.repo.List(sekolahID, params)
}

func (s *Service) GetByID(sekolahID, id int64) (*User, error) {
	user, err := s.repo.GetByID(sekolahID, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user tidak ditemukan")
		}
		return nil, err
	}
	return user, nil
}

func (s *Service) Create(sekolahID int64, req CreateRequest) (*User, error) {
	if !isValidRole(req.Role) {
		return nil, fmt.Errorf("role tidak valid")
	}

	exists, err := s.repo.EmailExists(sekolahID, req.Email, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("email sudah digunakan")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &User{
		SekolahID: sekolahID,
		Email:     req.Email,
		Password:  string(hashed),
		Nama:      req.Nama,
		Role:      req.Role,
		NoHP:      req.NoHP,
		Aktif:     true,
	}

	id, err := s.repo.Create(u)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(sekolahID, id)
}

func (s *Service) Update(sekolahID, id int64, req UpdateRequest) (*User, error) {
	if !isValidRole(req.Role) {
		return nil, fmt.Errorf("role tidak valid")
	}

	existing, err := s.repo.GetByID(sekolahID, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user tidak ditemukan")
		}
		return nil, err
	}

	exists, err := s.repo.EmailExists(sekolahID, req.Email, &id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("email sudah digunakan")
	}

	if existing.Role == "admin" && existing.Aktif && (!req.Aktif || req.Role != "admin") {
		count, err := s.repo.CountActiveAdmins(sekolahID)
		if err != nil {
			return nil, err
		}
		if count <= 1 {
			return nil, fmt.Errorf("tidak dapat menonaktifkan admin terakhir")
		}
	}

	if err := s.repo.Update(sekolahID, id, req.Nama, req.Email, req.Role, req.NoHP, req.Aktif); err != nil {
		return nil, err
	}

	return s.repo.GetByID(sekolahID, id)
}

func (s *Service) Deactivate(sekolahID, id int64) error {
	existing, err := s.repo.GetByID(sekolahID, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user tidak ditemukan")
		}
		return err
	}

	if existing.Role == "admin" && existing.Aktif {
		count, err := s.repo.CountActiveAdmins(sekolahID)
		if err != nil {
			return err
		}
		if count <= 1 {
			return fmt.Errorf("tidak dapat menonaktifkan admin terakhir")
		}
	}

	return s.repo.Update(sekolahID, id, existing.Nama, existing.Email, existing.Role, existing.NoHP, false)
}

func (s *Service) ResetPassword(sekolahID, id int64, req ResetPasswordRequest) error {
	_, err := s.repo.GetByID(sekolahID, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user tidak ditemukan")
		}
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.repo.UpdatePassword(sekolahID, id, string(hashed))
}

func isValidRole(role string) bool {
	for _, r := range validRoles {
		if r == role {
			return true
		}
	}
	return false
}
