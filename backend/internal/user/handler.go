package user

import (
	"net/http"
	"strconv"

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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	params := ListParams{
		Page:   page,
		Limit:  limit,
		Search: r.URL.Query().Get("search"),
		Role:   r.URL.Query().Get("role"),
	}

	if aktifStr := r.URL.Query().Get("aktif"); aktifStr != "" {
		aktif := aktifStr == "true"
		params.Aktif = &aktif
	}

	users, total, err := h.service.List(user.SekolahID, params)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data pengguna")
		return
	}

	if users == nil {
		users = []User{}
	}

	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	response.JSONWithMeta(w, 200, users, &response.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "ID tidak valid")
		return
	}

	result, err := h.service.GetByID(user.SekolahID, id)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", err.Error())
		return
	}

	response.JSON(w, 200, result)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	var req CreateRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	errs := validator.Collect(
		validator.Required("nama", req.Nama),
		validator.Required("email", req.Email),
		validator.Email("email", req.Email),
		validator.Required("password", req.Password),
		validator.MinLength("password", req.Password, 8),
		validator.Required("role", req.Role),
		validator.InList("role", req.Role, validRoles),
	)
	if len(errs) > 0 {
		response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", errs)
		return
	}

	result, err := h.service.Create(user.SekolahID, req)
	if err != nil {
		response.Error(w, 400, "CREATE_FAILED", err.Error())
		return
	}

	response.JSON(w, 201, result)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "ID tidak valid")
		return
	}

	var req UpdateRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	errs := validator.Collect(
		validator.Required("nama", req.Nama),
		validator.Required("email", req.Email),
		validator.Email("email", req.Email),
		validator.Required("role", req.Role),
		validator.InList("role", req.Role, validRoles),
	)
	if len(errs) > 0 {
		response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", errs)
		return
	}

	result, err := h.service.Update(user.SekolahID, id, req)
	if err != nil {
		response.Error(w, 400, "UPDATE_FAILED", err.Error())
		return
	}

	response.JSON(w, 200, result)
}

func (h *Handler) Deactivate(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "ID tidak valid")
		return
	}

	if err := h.service.Deactivate(user.SekolahID, id); err != nil {
		response.Error(w, 400, "DEACTIVATE_FAILED", err.Error())
		return
	}

	response.JSON(w, 200, map[string]string{"message": "User berhasil dinonaktifkan"})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "ID tidak valid")
		return
	}

	var req ResetPasswordRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	errs := validator.Collect(
		validator.Required("password", req.Password),
		validator.MinLength("password", req.Password, 8),
	)
	if len(errs) > 0 {
		response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", errs)
		return
	}

	if err := h.service.ResetPassword(user.SekolahID, id, req); err != nil {
		response.Error(w, 400, "RESET_FAILED", err.Error())
		return
	}

	response.JSON(w, 200, map[string]string{"message": "Password berhasil direset"})
}
