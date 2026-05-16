package upload

import (
	"net/http"
	"strings"

	"github.com/Sekolahkit/sekolah-app/pkg/middleware"
	"github.com/Sekolahkit/sekolah-app/pkg/response"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
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
