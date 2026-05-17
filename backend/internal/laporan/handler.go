package laporan

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Sekolahkit/sekolah-app/pkg/export"
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

func (h *Handler) ExportPembayaran(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	tanggalMulai := r.URL.Query().Get("tanggal_mulai")
	tanggalSelesai := r.URL.Query().Get("tanggal_selesai")
	if tanggalMulai == "" || tanggalSelesai == "" {
		response.Error(w, 400, "INVALID_REQUEST", "Parameter tanggal_mulai dan tanggal_selesai wajib diisi")
		return
	}

	tahunAjaranID, _ := strconv.ParseInt(r.URL.Query().Get("tahun_ajaran_id"), 10, 64)

	list, err := h.service.ExportPembayaran(user.SekolahID, tanggalMulai, tanggalSelesai, tahunAjaranID)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data pembayaran")
		return
	}

	cols := []export.Column{
		{Header: "Tanggal", Width: 14},
		{Header: "Siswa", Width: 25},
		{Header: "Kategori", Width: 18},
		{Header: "Metode", Width: 14},
		{Header: "Jumlah", Width: 16},
		{Header: "Status", Width: 12},
	}

	var rows [][]string
	for _, e := range list {
		rows = append(rows, []string{
			e.Tanggal, e.SiswaNama, e.Kategori, e.Metode,
			fmt.Sprintf("%.0f", e.Jumlah), e.Status,
		})
	}

	if err := export.WriteXLSX(w, "laporan-pembayaran.xlsx", cols, rows); err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal membuat file export")
	}
}

func (h *Handler) ExportPPDB(w http.ResponseWriter, r *http.Request) {
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

	list, err := h.service.ExportPPDB(user.SekolahID, tahunAjaranID)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data PPDB")
		return
	}

	cols := []export.Column{
		{Header: "Nama Lengkap", Width: 25},
		{Header: "NIK", Width: 20},
		{Header: "Jenis Kelamin", Width: 14},
		{Header: "Asal Sekolah", Width: 22},
		{Header: "Status", Width: 14},
		{Header: "Skor", Width: 10},
		{Header: "Ranking", Width: 10},
	}

	var rows [][]string
	for _, e := range list {
		rows = append(rows, []string{
			e.NamaLengkap, e.NIK, e.JenisKelamin, e.AsalSekolah,
			e.Status, fmt.Sprintf("%.2f", e.Skor), strconv.Itoa(e.Ranking),
		})
	}

	if err := export.WriteXLSX(w, "laporan-ppdb.xlsx", cols, rows); err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal membuat file export")
	}
}

func (h *Handler) ExportSiswa(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	tahunAjaranID, _ := strconv.ParseInt(r.URL.Query().Get("tahun_ajaran_id"), 10, 64)

	list, err := h.service.ExportSiswa(user.SekolahID, tahunAjaranID)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data siswa")
		return
	}

	cols := []export.Column{
		{Header: "NIS", Width: 12},
		{Header: "Nama", Width: 25},
		{Header: "Jenis Kelamin", Width: 14},
		{Header: "Tempat Lahir", Width: 16},
		{Header: "Tanggal Lahir", Width: 14},
		{Header: "Agama", Width: 10},
		{Header: "Alamat", Width: 30},
		{Header: "Status", Width: 10},
	}

	var rows [][]string
	for _, e := range list {
		rows = append(rows, []string{
			e.NIS, e.Nama, e.JenisKelamin, e.TempatLahir,
			e.TanggalLahir, e.Agama, e.Alamat, e.Status,
		})
	}

	if err := export.WriteXLSX(w, "laporan-siswa.xlsx", cols, rows); err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal membuat file export")
	}
}
