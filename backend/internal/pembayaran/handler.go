package pembayaran

import (
	"fmt"
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

func (h *Handler) ListKategori(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	list, err := h.service.ListKategori(user.SekolahID)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data kategori")
		return
	}

	response.JSON(w, 200, list)
}

func (h *Handler) GetKategoriByID(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	k, err := h.service.GetKategoriByID(user.SekolahID, id)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", "Kategori tidak ditemukan")
		return
	}

	response.JSON(w, 200, k)
}

func (h *Handler) CreateKategori(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	var req CreateKategoriRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	k, err := h.service.CreateKategori(user.SekolahID, req)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, 201, k)
}

func (h *Handler) UpdateKategori(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	var req UpdateKategoriRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	k, err := h.service.UpdateKategori(user.SekolahID, id, req)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, 200, k)
}

func (h *Handler) DeleteKategori(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	if err := h.service.DeleteKategori(user.SekolahID, id); err != nil {
		response.Error(w, 404, "NOT_FOUND", "Kategori tidak ditemukan")
		return
	}

	response.JSON(w, 200, map[string]string{"message": "Kategori berhasil dihapus"})
}

func (h *Handler) ListTagihan(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	params := parseTagihanListParams(r)

	list, total, err := h.service.ListTagihan(user.SekolahID, params)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data tagihan")
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

func (h *Handler) GetTagihanByID(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	t, err := h.service.GetTagihanByID(user.SekolahID, id)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", "Tagihan tidak ditemukan")
		return
	}

	response.JSON(w, 200, t)
}

func (h *Handler) CreateTagihan(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	var req CreateTagihanRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	t, err := h.service.CreateTagihan(user.SekolahID, req)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, 201, t)
}

func (h *Handler) BulkCreateTagihan(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	var req BulkCreateTagihanRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	if err := h.service.BulkCreateTagihan(user.SekolahID, req); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, 201, map[string]string{"message": "Tagihan berhasil dibuat"})
}

func (h *Handler) UpdateTagihan(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	var req UpdateTagihanRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	t, err := h.service.UpdateTagihan(user.SekolahID, id, req)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, 200, t)
}

func (h *Handler) DeleteTagihan(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	if err := h.service.DeleteTagihan(user.SekolahID, id); err != nil {
		response.Error(w, 404, "NOT_FOUND", "Tagihan tidak ditemukan")
		return
	}

	response.JSON(w, 200, map[string]string{"message": "Tagihan berhasil dihapus"})
}

func (h *Handler) ListPembayaran(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	params := parsePembayaranListParams(r)

	list, total, err := h.service.ListPembayaran(user.SekolahID, params)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data pembayaran")
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

func (h *Handler) GetPembayaranByID(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	p, err := h.service.GetPembayaranByID(user.SekolahID, id)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", "Pembayaran tidak ditemukan")
		return
	}

	response.JSON(w, 200, p)
}

func (h *Handler) CreatePembayaran(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	var req CreatePembayaranRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	p, err := h.service.CreatePembayaran(user.SekolahID, req)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, 201, p)
}

func (h *Handler) VerifyPembayaran(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	if err := h.service.VerifyPembayaran(user.SekolahID, id, user.UserID); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", ve)
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, 200, map[string]string{"message": "Pembayaran berhasil diverifikasi"})
}

func (h *Handler) RejectPembayaran(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	if err := h.service.RejectPembayaran(user.SekolahID, id); err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, 200, map[string]string{"message": "Pembayaran berhasil ditolak"})
}

func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	params := parseTagihanListParams(r)

	list, err := h.service.ExportTagihan(user.SekolahID, params)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data tagihan")
		return
	}

	cols := []export.Column{
		{Header: "NIS", Width: 12},
		{Header: "Nama Siswa", Width: 25},
		{Header: "Kategori", Width: 18},
		{Header: "Tahun Ajaran", Width: 16},
		{Header: "Semester", Width: 10},
		{Header: "Nominal", Width: 16},
		{Header: "Jatuh Tempo", Width: 14},
		{Header: "Status", Width: 12},
	}

	var rows [][]string
	for _, t := range list {
		rows = append(rows, []string{
			t.SiswaNIS, t.SiswaNama, t.KategoriNama, t.TahunAjaranNama,
			t.Semester, fmt.Sprintf("%.0f", t.Nominal), t.JatuhTempo, t.Status,
		})
	}

	if err := export.WriteXLSX(w, "data-tagihan.xlsx", cols, rows); err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal membuat file export")
	}
}

func (h *Handler) Kwitansi(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_ID", "ID tidak valid")
		return
	}

	data, err := h.service.GetKwitansiData(user.SekolahID, id)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", "Pembayaran tidak ditemukan")
		return
	}

	statusLabel := map[string]string{
		"verified": "Terverifikasi",
		"pending":  "Menunggu Verifikasi",
		"rejected": "Ditolak",
	}
	metodeLabel := map[string]string{
		"transfer": "Transfer Bank",
		"cash":     "Tunai",
		"midtrans":  "Midtrans",
		"xendit":   "Xendit",
	}

	catatanHTML := ""
	if data.Catatan != "" {
		catatanHTML = fmt.Sprintf(`<div class="row"><span class="label">Catatan</span><span class="value">%s</span></div>`, data.Catatan)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, kwitansiHTML,
		data.SekolahNama,
		data.SiswaNama,
		data.SiswaNIS,
		data.KategoriNama,
		data.TahunAjaranNama,
		formatRupiah(data.NominalTagihan),
		formatRupiah(data.JumlahBayar),
		data.TanggalBayar,
		metodeLabel[data.MetodeBayar],
		statusLabel[data.StatusBayar],
		data.VerifiedAt,
		catatanHTML,
	)
}

func formatRupiah(v float64) string {
	s := fmt.Sprintf("%.0f", v)
	n := len(s)
	if n <= 3 {
		return "Rp " + s
	}
	result := ""
	for i, c := range s {
		if i > 0 && (n-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}
	return "Rp " + result
}

const kwitansiHTML = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Kwitansi Pembayaran</title>
<style>
  @media print { body { margin: 0; } @page { margin: 2cm; } }
  body { font-family: 'Segoe UI', Arial, sans-serif; max-width: 600px; margin: 40px auto; color: #333; }
  .header { text-align: center; border-bottom: 3px double #333; padding-bottom: 16px; margin-bottom: 20px; }
  .header h1 { margin: 0; font-size: 20px; text-transform: uppercase; }
  .header p { margin: 4px 0 0; font-size: 13px; color: #666; }
  .row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #eee; }
  .label { color: #666; font-size: 14px; min-width: 160px; }
  .value { font-size: 14px; font-weight: 500; text-align: right; }
  .amount { font-size: 24px; font-weight: 700; text-align: center; margin: 24px 0; padding: 16px; background: #f8f9fa; border: 1px solid #dee2e6; border-radius: 8px; }
  .amount small { display: block; font-size: 12px; color: #666; font-weight: 400; margin-bottom: 4px; }
  .footer { text-align: center; margin-top: 40px; font-size: 12px; color: #999; }
</style></head><body>
<div class="header">
  <h1>%s</h1>
  <p>Kwitansi Pembayaran</p>
</div>
<div class="row"><span class="label">Siswa</span><span class="value">%s (%s)</span></div>
<div class="row"><span class="label">Kategori</span><span class="value">%s</span></div>
<div class="row"><span class="label">Tahun Ajaran</span><span class="value">%s</span></div>
<div class="row"><span class="label">Nominal Tagihan</span><span class="value">%s</span></div>
<div class="amount"><small>Jumlah Dibayar</small>%s</div>
<div class="row"><span class="label">Tanggal</span><span class="value">%s</span></div>
<div class="row"><span class="label">Metode</span><span class="value">%s</span></div>
<div class="row"><span class="label">Status</span><span class="value">%s</span></div>
<div class="row"><span class="label">Diverifikasi</span><span class="value">%s</span></div>
%s
<div class="footer">Dicetak dari sistem sekolah</div>
<script>window.onload=function(){window.print()}</script>
</body></html>`

func parseTagihanListParams(r *http.Request) TagihanListParams {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	sort := r.URL.Query().Get("sort")
	allowedSorts := map[string]bool{"nominal": true, "jatuh_tempo": true, "status": true, "created_at": true}
	if !allowedSorts[sort] {
		sort = "created_at"
	}
	order := r.URL.Query().Get("order")
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	siswaID, _ := strconv.ParseInt(r.URL.Query().Get("siswa_id"), 10, 64)
	kategoriID, _ := strconv.ParseInt(r.URL.Query().Get("kategori_id"), 10, 64)
	tahunAjaranID, _ := strconv.ParseInt(r.URL.Query().Get("tahun_ajaran_id"), 10, 64)
	return TagihanListParams{
		Page:          page,
		Limit:         limit,
		Sort:          sort,
		Order:         order,
		SiswaID:       siswaID,
		KategoriID:    kategoriID,
		TahunAjaranID: tahunAjaranID,
		Status:        r.URL.Query().Get("status"),
	}
}

func parsePembayaranListParams(r *http.Request) PembayaranListParams {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	sort := r.URL.Query().Get("sort")
	allowedSorts := map[string]bool{"jumlah": true, "tanggal": true, "status": true, "created_at": true}
	if !allowedSorts[sort] {
		sort = "created_at"
	}
	order := r.URL.Query().Get("order")
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	tagihanID, _ := strconv.ParseInt(r.URL.Query().Get("tagihan_id"), 10, 64)
	return PembayaranListParams{
		Page:      page,
		Limit:     limit,
		Sort:      sort,
		Order:     order,
		TagihanID: tagihanID,
		Status:    r.URL.Query().Get("status"),
	}
}
