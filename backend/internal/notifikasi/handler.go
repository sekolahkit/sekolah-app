package notifikasi

import (
	"net/http"
	"strconv"

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

type TestSendRequest struct {
	Tipe     string `json:"tipe"`
	Penerima string `json:"penerima"`
	Pesan    string `json:"pesan"`
}

func (h *Handler) ListNotifikasi(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	params := parseListParams(r)

	list, total, err := h.service.List(user.SekolahID, params)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data notifikasi")
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

func (h *Handler) TestSend(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	var req TestSendRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	n, err := h.service.TestSend(user.SekolahID, req.Tipe, req.Penerima, req.Pesan)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengirim notifikasi")
		return
	}

	response.JSON(w, 201, n)
}

func (h *Handler) QueueStatus(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	stats, err := h.service.GetQueueStats(user.SekolahID)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil status antrian")
		return
	}

	response.JSON(w, 200, stats)
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
	return ListParams{
		Page:   page,
		Limit:  limit,
		Status: r.URL.Query().Get("status"),
		Tipe:   r.URL.Query().Get("tipe"),
	}
}
