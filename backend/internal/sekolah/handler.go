package sekolah

import (
	"net/http"

	"github.com/Sekolahkit/sekolah-app/pkg/middleware"
	"github.com/Sekolahkit/sekolah-app/pkg/response"
	"github.com/Sekolahkit/sekolah-app/pkg/validator"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	sek, err := h.service.Get(user.SekolahID)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", "Sekolah tidak ditemukan")
		return
	}

	response.JSON(w, 200, sek)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	var req UpdateRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	sek, err := h.service.Update(user.SekolahID, req)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, 200, sek)
}
