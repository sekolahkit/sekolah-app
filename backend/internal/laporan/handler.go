package laporan

import (
	"net/http"
	"strconv"

	"github.com/Sekolahkit/sekolah-app/pkg/middleware"
	"github.com/Sekolahkit/sekolah-app/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) DashboardAdmin(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	stats, err := h.service.GetDashboard(user.SekolahID)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data dashboard")
		return
	}

	response.JSON(w, 200, stats)
}

func (h *Handler) DashboardOperator(w http.ResponseWriter, r *http.Request) {
	h.DashboardAdmin(w, r)
}

func (h *Handler) RekapPembayaran(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	tanggalMulai := r.URL.Query().Get("tanggal_mulai")
	tanggalSelesai := r.URL.Query().Get("tanggal_selesai")

	if tanggalMulai == "" || tanggalSelesai == "" {
		response.Error(w, 400, "INVALID_REQUEST", "Parameter tanggal_mulai dan tanggal_selesai wajib diisi")
		return
	}

	tahunAjaranID, _ := strconv.ParseInt(r.URL.Query().Get("tahun_ajaran_id"), 10, 64)

	data, err := h.service.GetRekapPembayaran(user.SekolahID, tanggalMulai, tanggalSelesai, tahunAjaranID)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil rekap pembayaran")
		return
	}

	response.JSON(w, 200, data)
}

func (h *Handler) RekapPPDB(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	tahunAjaranIDStr := r.URL.Query().Get("tahun_ajaran_id")
	if tahunAjaranIDStr == "" {
		response.Error(w, 400, "INVALID_REQUEST", "Parameter tahun_ajaran_id wajib diisi")
		return
	}

	tahunAjaranID, err := strconv.ParseInt(tahunAjaranIDStr, 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Parameter tahun_ajaran_id tidak valid")
		return
	}

	data, err := h.service.GetRekapPPDB(user.SekolahID, tahunAjaranID)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil rekap PPDB")
		return
	}

	response.JSON(w, 200, data)
}

func (h *Handler) RekapSiswa(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	tahunAjaranID, _ := strconv.ParseInt(r.URL.Query().Get("tahun_ajaran_id"), 10, 64)

	data, err := h.service.GetRekapSiswa(user.SekolahID, tahunAjaranID)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil rekap siswa")
		return
	}

	response.JSON(w, 200, data)
}
