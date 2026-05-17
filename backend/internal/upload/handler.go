package upload

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Sekolahkit/sekolah-app/pkg/middleware"
	"github.com/Sekolahkit/sekolah-app/pkg/response"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
	secret  string
}

func NewHandler(service *Service, secret string) *Handler {
	return &Handler{service: service, secret: secret}
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		response.Error(w, 401, "UNAUTHORIZED", "Belum login")
		return
	}

	if err := r.ParseMultipartForm(h.service.maxSize); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "File terlalu besar atau request tidak valid")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "File tidak ditemukan dalam request")
		return
	}
	defer file.Close()

	category := r.FormValue("category")
	if category == "" {
		category = "general"
	}
	allowedCategories := []string{"bukti_bayar", "berkas_ppdb", "foto_siswa", "foto_ppdb", "general"}
	validCategory := false
	for _, c := range allowedCategories {
		if category == c {
			validCategory = true
			break
		}
	}
	if !validCategory {
		response.Error(w, 400, "INVALID_REQUEST", "Kategori upload tidak valid")
		return
	}

	path, err := h.service.Upload(user.SekolahID, category, file, header)
	if err != nil {
		response.Error(w, 400, "UPLOAD_FAILED", err.Error())
		return
	}

	response.JSON(w, 201, map[string]string{"path": path})
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		response.Error(w, 401, "UNAUTHORIZED", "Belum login")
		return
	}

	path := chi.URLParam(r, "*")
	if path == "" {
		response.Error(w, 400, "INVALID_REQUEST", "Path tidak valid")
		return
	}

	path = strings.TrimPrefix(path, "/")

	if err := h.service.ValidateAccess(path, user.SekolahID); err != nil {
		response.Error(w, 404, "NOT_FOUND", "File tidak ditemukan")
		return
	}

	fullPath, err := h.service.GetFilePath(path)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", "File tidak ditemukan")
		return
	}

	http.ServeFile(w, r, fullPath)
}

type generateSignedRequest struct {
	Path      string `json:"path"`
	TTLSeconds int   `json:"ttl_seconds"`
}

func (h *Handler) GenerateSignedURL(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		response.Error(w, 401, "UNAUTHORIZED", "Belum login")
		return
	}

	var req generateSignedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	if req.Path == "" {
		response.Error(w, 400, "INVALID_REQUEST", "Path tidak boleh kosong")
		return
	}

	if err := ValidatePath(req.Path); err != nil {
		response.Error(w, 403, "FORBIDDEN", "Path tidak valid")
		return
	}

	if err := h.service.ValidateAccess(req.Path, user.SekolahID); err != nil {
		response.Error(w, 403, "FORBIDDEN", "Tidak punya akses ke file ini")
		return
	}

	if _, err := h.service.GetFilePath(req.Path); err != nil {
		response.Error(w, 404, "NOT_FOUND", "File tidak ditemukan")
		return
	}

	ttl := time.Duration(req.TTLSeconds) * time.Second
	result, err := SignPath(h.secret, user.SekolahID, req.Path, ttl)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal membuat signed URL")
		return
	}

	response.JSON(w, 200, result)
}

func (h *Handler) ServeSignedURL(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		response.Error(w, 400, "INVALID_REQUEST", "Token tidak valid")
		return
	}

	payload, err := ValidateSignedToken(h.secret, token)
	if err != nil {
		response.Error(w, 403, "FORBIDDEN", err.Error())
		return
	}

	fullPath, err := h.service.GetFilePath(payload.Path)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", "File tidak ditemukan")
		return
	}

	http.ServeFile(w, r, fullPath)
}
