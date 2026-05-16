package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo      *Repository
	jwtSecret string
}

func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret}
}

type LoginRequest struct {
	KodeSekolah string `json:"kode_sekolah"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}

type LoginResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"-"`
	RefreshToken string        `json:"-"`
}

type UserResponse struct {
	ID        int64  `json:"id"`
	SekolahID int64  `json:"sekolah_id"`
	Email     string `json:"email"`
	Nama      string `json:"nama"`
	Role      string `json:"role"`
}

func (s *Service) Login(req LoginRequest, ipAddress, userAgent string) (*LoginResponse, error) {
	sekolahID, err := s.repo.FindSekolahByKode(req.KodeSekolah)
	if err != nil {
		if err == sql.ErrNoRows {
			s.repo.RecordLoginAttempt(nil, req.Email, ipAddress, false)
			return nil, fmt.Errorf("kode sekolah atau email/password salah")
		}
		return nil, fmt.Errorf("find sekolah: %w", err)
	}

	lockoutDuration := s.checkLockout(&sekolahID, req.Email)
	if lockoutDuration > 0 {
		return nil, fmt.Errorf("akun terkunci, coba lagi dalam %d menit", int(lockoutDuration.Minutes())+1)
	}

	user, err := s.repo.FindUserByEmail(sekolahID, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			s.repo.RecordLoginAttempt(&sekolahID, req.Email, ipAddress, false)
			return nil, fmt.Errorf("kode sekolah atau email/password salah")
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if !user.Aktif {
		return nil, fmt.Errorf("akun tidak aktif")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.repo.RecordLoginAttempt(&sekolahID, req.Email, ipAddress, false)
		return nil, fmt.Errorf("kode sekolah atau email/password salah")
	}

	s.repo.RecordLoginAttempt(&sekolahID, req.Email, ipAddress, true)

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.generateAndSaveRefreshToken(user.ID, userAgent)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &LoginResponse{
		User: &UserResponse{
			ID:        user.ID,
			SekolahID: user.SekolahID,
			Email:     user.Email,
			Nama:      user.Nama,
			Role:      user.Role,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) Refresh(refreshTokenRaw, userAgent string) (*LoginResponse, error) {
	tokenHash := HashToken(refreshTokenRaw)

	rt, err := s.repo.FindRefreshToken(tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("refresh token tidak valid")
		}
		return nil, fmt.Errorf("find refresh token: %w", err)
	}

	if rt.RevokedAt.Valid {
		s.repo.RevokeAllUserTokens(rt.PenggunaID)
		return nil, fmt.Errorf("refresh token sudah digunakan, semua session di-revoke")
	}

	if time.Now().After(rt.ExpiresAt) {
		return nil, fmt.Errorf("refresh token expired")
	}

	s.repo.RevokeRefreshToken(rt.ID)

	user, err := s.repo.FindUserByID(rt.PenggunaID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	if !user.Aktif {
		return nil, fmt.Errorf("akun tidak aktif")
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, err := s.generateAndSaveRefreshToken(user.ID, userAgent)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &LoginResponse{
		User: &UserResponse{
			ID:        user.ID,
			SekolahID: user.SekolahID,
			Email:     user.Email,
			Nama:      user.Nama,
			Role:      user.Role,
		},
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *Service) Logout(refreshTokenRaw string) error {
	tokenHash := HashToken(refreshTokenRaw)
	rt, err := s.repo.FindRefreshToken(tokenHash)
	if err != nil {
		return nil
	}
	return s.repo.RevokeRefreshToken(rt.ID)
}

func (s *Service) RevokeAll(penggunaID int64) error {
	return s.repo.RevokeAllUserTokens(penggunaID)
}

func (s *Service) GetCurrentUser(userID int64) (*UserResponse, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, err
	}
	return &UserResponse{
		ID:        user.ID,
		SekolahID: user.SekolahID,
		Email:     user.Email,
		Nama:      user.Nama,
		Role:      user.Role,
	}, nil
}

func (s *Service) generateAccessToken(user *User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    user.ID,
		"sekolah_id": user.SekolahID,
		"role":       user.Role,
		"email":      user.Email,
		"exp":        time.Now().Add(15 * time.Minute).Unix(),
		"iat":        time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *Service) generateAndSaveRefreshToken(penggunaID int64, deviceInfo string) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	tokenStr := hex.EncodeToString(raw)
	tokenHash := HashToken(tokenStr)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	if err := s.repo.SaveRefreshToken(penggunaID, tokenHash, deviceInfo, expiresAt); err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (s *Service) checkLockout(sekolahID *int64, email string) time.Duration {
	since := time.Now().Add(-30 * time.Minute)
	count, err := s.repo.CountRecentFailedAttempts(sekolahID, email, since)
	if err != nil {
		return 0
	}

	switch {
	case count >= 15:
		return 24 * time.Hour
	case count >= 10:
		return 30 * time.Minute
	case count >= 5:
		return 5 * time.Minute
	case count >= 3:
		return 30 * time.Second
	default:
		return 0
	}
}
