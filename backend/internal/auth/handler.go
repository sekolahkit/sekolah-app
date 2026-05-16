package auth

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Sekolahkit/sekolah-app/pkg/middleware"
	"github.com/Sekolahkit/sekolah-app/pkg/response"
	"github.com/Sekolahkit/sekolah-app/pkg/validator"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	errs := validator.Collect(
		validator.Required("kode_sekolah", req.KodeSekolah),
		validator.Required("email", req.Email),
		validator.Required("password", req.Password),
		validator.Email("email", req.Email),
	)
	if len(errs) > 0 {
		response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", errs)
		return
	}

	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()

	result, err := h.service.Login(req, ipAddress, userAgent)
	if err != nil {
		response.Error(w, 401, "LOGIN_FAILED", err.Error())
		return
	}

	setAccessTokenCookie(w, result.AccessToken)
	setRefreshTokenCookie(w, result.RefreshToken)

	response.JSON(w, 200, result.User)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		response.Error(w, 401, "UNAUTHORIZED", "Refresh token tidak ditemukan")
		return
	}

	userAgent := r.UserAgent()

	result, err := h.service.Refresh(cookie.Value, userAgent)
	if err != nil {
		clearAuthCookies(w)
		response.Error(w, 401, "REFRESH_FAILED", err.Error())
		return
	}

	setAccessTokenCookie(w, result.AccessToken)
	setRefreshTokenCookie(w, result.RefreshToken)

	response.JSON(w, 200, result.User)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		h.service.Logout(cookie.Value)
	}

	clearAuthCookies(w)
	response.JSON(w, 200, map[string]string{"message": "Berhasil logout"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		response.Error(w, 401, "UNAUTHORIZED", "Belum login")
		return
	}

	result, err := h.service.GetCurrentUser(user.UserID)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", "User tidak ditemukan")
		return
	}

	response.JSON(w, 200, result)
}

func (h *Handler) RevokeAll(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "User ID tidak valid")
		return
	}

	if err := h.service.RevokeAll(userID); err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal revoke session")
		return
	}

	response.JSON(w, 200, map[string]string{"message": "Semua session berhasil di-revoke"})
}

func setAccessTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		MaxAge:   900,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})
}

func setRefreshTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   int(7 * 24 * time.Hour / time.Second),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})
}

func clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth/refresh",
		MaxAge:   -1,
		HttpOnly: true,
	})
}
