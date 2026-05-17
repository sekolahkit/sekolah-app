package backup

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	backups, err := h.service.List()
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil daftar backup")
		return
	}

	response.JSON(w, 200, backups)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Cleanup(); err != nil {
		fmt.Printf("warning: cleanup gagal: %v\n", err)
	}

	info, err := h.service.Create()
	if err != nil {
		response.Error(w, 500, "BACKUP_FAILED", err.Error())
		return
	}

	response.JSON(w, 201, info)
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, 400, "INVALID_REQUEST", "ID backup diperlukan")
		return
	}

	if strings.Contains(id, "..") || strings.Contains(id, "/") {
		response.Error(w, 400, "INVALID_REQUEST", "ID backup tidak valid")
		return
	}

	filePath, filename, err := h.service.GetFilePath(id)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", "Backup tidak ditemukan")
		return
	}

	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(filename)))
	http.ServeFile(w, r, filePath)
}

type RestoreRequest struct {
	Confirm  string `json:"confirm"`
	BackupID string `json:"backup_id"`
}

func (h *Handler) Restore(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, 400, "INVALID_REQUEST", "ID backup diperlukan")
		return
	}

	if strings.Contains(id, "..") || strings.Contains(id, "/") {
		response.Error(w, 400, "INVALID_REQUEST", "ID backup tidak valid")
		return
	}

	var req RestoreRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	errs := validator.Collect(
		validator.Required("confirm", req.Confirm),
		validator.Required("backup_id", req.BackupID),
	)
	if len(errs) > 0 {
		response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", errs)
		return
	}

	if req.Confirm != "RESTORE" {
		response.Error(w, 400, "CONFIRMATION_REQUIRED", "Ketik 'RESTORE' untuk konfirmasi")
		return
	}

	if req.BackupID != id {
		response.Error(w, 400, "INVALID_REQUEST", "Backup ID tidak cocok")
		return
	}

	if err := h.service.Restore(id); err != nil {
		response.Error(w, 500, "RESTORE_FAILED", err.Error())
		return
	}

	response.JSON(w, 200, map[string]string{"message": "Restore berhasil. Silakan restart aplikasi."})
}
