package ppdb

import (
	"net/http"
	"strconv"

	"github.com/Sekolahkit/sekolah-app/pkg/export"
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
		DaftarUlang:   r.URL.Query().Get("daftar_ulang"),
	}
}

func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	params := parseListParams(r)

	list, err := h.service.ExportPendaftar(user.SekolahID, params)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data pendaftar")
		return
	}

	cols := []export.Column{
		{Header: "No", Width: 6},
		{Header: "Nama Lengkap", Width: 25},
		{Header: "NIK", Width: 20},
		{Header: "Jenis Kelamin", Width: 14},
		{Header: "Tempat, Tanggal Lahir", Width: 22},
		{Header: "Agama", Width: 10},
		{Header: "Alamat", Width: 30},
		{Header: "Asal Sekolah", Width: 22},
		{Header: "No HP", Width: 14},
		{Header: "Email", Width: 22},
		{Header: "Nama Ortu", Width: 22},
		{Header: "No HP Ortu", Width: 14},
		{Header: "Pekerjaan Ortu", Width: 18},
		{Header: "Status", Width: 14},
	}

	var rows [][]string
	for i, p := range list {
		ttl := p.TempatLahir
		if p.TanggalLahir != "" {
			if ttl != "" {
				ttl += ", "
			}
			ttl += p.TanggalLahir
		}
		rows = append(rows, []string{
			strconv.Itoa(i + 1), p.NamaLengkap, p.NIK, p.JenisKelamin,
			ttl, p.Agama, p.Alamat, p.AsalSekolah, p.NoHP, p.Email,
			p.NamaOrtu, p.NoHPOrtu, p.PekerjaanOrtu, p.Status,
		})
	}

	if err := export.WriteXLSX(w, "data-ppdb.xlsx", cols, rows); err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal membuat file export")
	}
}

func (h *Handler) RunRanking(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	var req RunRankingRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	result, err := h.service.RunRanking(user.SekolahID, user.UserID, req)
	if err != nil {
		response.Error(w, 400, "RANKING_ERROR", err.Error())
		return
	}

	response.JSON(w, 200, result)
}

func (h *Handler) PublishRanking(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	var req PublishRankingRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	count, err := h.service.PublishRanking(user.SekolahID, req)
	if err != nil {
		response.Error(w, 400, "PUBLISH_ERROR", err.Error())
		return
	}

	response.JSON(w, 200, map[string]interface{}{
		"message":          "Pengumuman berhasil dipublish",
		"published_count":  count,
	})
}

func (h *Handler) DaftarUlang(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	if err := h.service.DaftarUlang(user.SekolahID, id); err != nil {
		response.Error(w, 400, "DAFTAR_ULANG_ERROR", err.Error())
		return
	}

	response.JSON(w, 200, map[string]string{"message": "Daftar ulang berhasil"})
}

func (h *Handler) GetDaftarUlangStatus(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	status, err := h.service.GetDaftarUlangStatus(user.SekolahID, id)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", err.Error())
		return
	}

	response.JSON(w, 200, status)
}
