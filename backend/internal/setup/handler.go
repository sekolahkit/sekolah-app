package setup

import (
	"net/http"

	"github.com/Sekolahkit/sekolah-app/pkg/response"
	"github.com/Sekolahkit/sekolah-app/pkg/validator"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	initialized, err := h.service.IsInitialized()
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal cek status setup")
		return
	}
	response.JSON(w, 200, map[string]bool{"initialized": initialized})
}

func (h *Handler) RunSetup(w http.ResponseWriter, r *http.Request) {
	var req SetupRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	errs := validator.Collect(
		validator.Required("nama_sekolah", req.NamaSekolah),
		validator.Required("kode_sekolah", req.KodeSekolah),
		validator.Required("nama_admin", req.NamaAdmin),
		validator.Required("email_admin", req.EmailAdmin),
		validator.Required("password_admin", req.PasswordAdmin),
		validator.Email("email_admin", req.EmailAdmin),
		validator.MinLength("password_admin", req.PasswordAdmin, 8),
		validator.MinLength("kode_sekolah", req.KodeSekolah, 3),
		validator.MaxLength("kode_sekolah", req.KodeSekolah, 30),
	)
	if len(errs) > 0 {
		response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", errs)
		return
	}

	result, err := h.service.RunSetup(req)
	if err != nil {
		response.Error(w, 422, "SETUP_FAILED", err.Error())
		return
	}

	response.JSON(w, 201, result)
}

func (h *Handler) SetupGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		initialized, err := h.service.IsInitialized()
		if err != nil {
			response.Error(w, 500, "INTERNAL_ERROR", "Gagal cek status setup")
			return
		}
		if initialized {
			response.Error(w, 404, "NOT_FOUND", "Endpoint tidak ditemukan")
			return
		}
		next.ServeHTTP(w, r)
	})
}
