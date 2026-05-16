package siswa

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
	params := parseListParams(r)

	list, total, err := h.service.List(user.SekolahID, params)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data siswa")
		return
	}

	totalPages := (total + params.Limit - 1) / params.Limit
	response.JSONWithMeta(w, 200, list, &response.Meta{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	siswa, err := h.service.GetByID(user.SekolahID, id)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", "Siswa tidak ditemukan")
		return
	}

	response.JSON(w, 200, siswa)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	var req CreateRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	siswa, err := h.service.Create(user.SekolahID, req)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, 201, siswa)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	var req UpdateRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	siswa, err := h.service.Update(user.SekolahID, id, req)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, 200, siswa)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	if err := h.service.Delete(user.SekolahID, id); err != nil {
		response.Error(w, 404, "NOT_FOUND", "Siswa tidak ditemukan")
		return
	}

	response.JSON(w, 200, map[string]string{"message": "Siswa berhasil dihapus"})
}

func (h *Handler) Import(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, 200, map[string]string{"message": "Import endpoint - coming soon"})
}

func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, 200, map[string]string{"message": "Export endpoint - coming soon"})
}

func parseListParams(r *http.Request) ListParams {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "created_at"
	}
	order := r.URL.Query().Get("order")
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	return ListParams{
		Page:   page,
		Limit:  limit,
		Search: r.URL.Query().Get("search"),
		Sort:   sort,
		Order:  order,
	}
}
