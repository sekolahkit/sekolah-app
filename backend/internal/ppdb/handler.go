package ppdb

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

func (h *Handler) Daftar(w http.ResponseWriter, r *http.Request) {
	var req DaftarRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	pendaftaran, err := h.service.Daftar(req)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, 201, pendaftaran)
}

func (h *Handler) GetPengumuman(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	pengumuman, err := h.service.GetPengumuman(id)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", "Pengumuman tidak ditemukan")
		return
	}

	response.JSON(w, 200, pengumuman)
}

func (h *Handler) ListPendaftar(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	params := parseListParams(r)

	list, total, err := h.service.ListPendaftar(user.SekolahID, params)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data pendaftar")
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

func (h *Handler) GetPendaftar(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	pendaftar, err := h.service.GetPendaftar(user.SekolahID, id)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", "Pendaftar tidak ditemukan")
		return
	}

	response.JSON(w, 200, pendaftar)
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	var req UpdateStatusRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	if err := h.service.UpdateStatus(user.SekolahID, id, req); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, 200, map[string]string{"message": "Status berhasil diperbarui"})
}

func (h *Handler) ListBerkas(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	berkas, err := h.service.ListBerkas(user.SekolahID, id)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", err.Error())
		return
	}

	response.JSON(w, 200, berkas)
}

func (h *Handler) VerifikasiBerkas(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	var req VerifikasiBerkasRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	if err := h.service.VerifikasiBerkas(user.SekolahID, id, req); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, 200, map[string]string{"message": "Berkas berhasil diverifikasi"})
}

func (h *Handler) InputUjian(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	var req InputUjianRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	ujian, err := h.service.InputNilaiUjian(user.SekolahID, req)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, 201, ujian)
}

func (h *Handler) PublishPengumuman(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	var req PublishPengumumanRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	pengumuman, err := h.service.PublishPengumuman(user.SekolahID, req)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, 201, pengumuman)
}

func (h *Handler) GetKonfigurasiRanking(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	tahunAjaranID, _ := strconv.ParseInt(r.URL.Query().Get("tahun_ajaran_id"), 10, 64)
	if tahunAjaranID == 0 {
		response.Error(w, 400, "INVALID_REQUEST", "tahun_ajaran_id wajib diisi")
		return
	}

	konfig, err := h.service.GetKonfigurasiRanking(user.SekolahID, tahunAjaranID)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", "Konfigurasi ranking tidak ditemukan")
		return
	}

	response.JSON(w, 200, konfig)
}

func (h *Handler) UpsertKonfigurasiRanking(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	var req KonfigurasiRankingRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	if err := h.service.UpsertKonfigurasiRanking(user.SekolahID, req); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, 200, map[string]string{"message": "Konfigurasi ranking berhasil disimpan"})
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
	tahunAjaranID, _ := strconv.ParseInt(r.URL.Query().Get("tahun_ajaran_id"), 10, 64)
	return ListParams{
		Page:          page,
		Limit:         limit,
		Status:        r.URL.Query().Get("status"),
		TahunAjaranID: tahunAjaranID,
	}
}
