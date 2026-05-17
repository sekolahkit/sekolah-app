package selfservice

import (
	"encoding/json"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Sekolahkit/sekolah-app/pkg/middleware"
	"github.com/Sekolahkit/sekolah-app/pkg/response"
	"github.com/Sekolahkit/sekolah-app/pkg/validator"
	"github.com/go-chi/chi/v5"
)

type GatewayPaymentResult struct {
	Provider        string `json:"provider"`
	OrderID         string `json:"order_id"`
	PaymentURL      string `json:"payment_url"`
	PaymentGatewayID string `json:"payment_gateway_id"`
	Status          string `json:"status"`
}

type GatewayPaymentInitiator interface {
	InitiateTransaction(sekolahID, tagihanID, createdBy int64, provider string) (*GatewayPaymentResult, error)
	HasProvider(provider string) bool
	AvailableProviders() []string
}

type Handler struct {
	service   *Service
	gateway   GatewayPaymentInitiator
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func NewHandlerWithGateway(service *Service, gateway GatewayPaymentInitiator) *Handler {
	return &Handler{service: service, gateway: gateway}
}

func (h *Handler) ListLinkedSiswa(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	result, err := h.service.GetLinkedSiswa(user.SekolahID, user.UserID)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data")
		return
	}
	if result == nil {
		result = []LinkedSiswa{}
	}
	response.JSON(w, 200, result)
}

func (h *Handler) GetSiswaDetail(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	siswaID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "ID tidak valid")
		return
	}

	result, err := h.service.GetSiswaDetail(user.SekolahID, user.UserID, siswaID)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", "Data tidak ditemukan")
		return
	}
	response.JSON(w, 200, result)
}

func (h *Handler) GetTagihan(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	siswaID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "ID tidak valid")
		return
	}

	result, err := h.service.GetTagihan(user.SekolahID, user.UserID, siswaID)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", "Data tidak ditemukan")
		return
	}
	if result == nil {
		result = []Tagihan{}
	}
	response.JSON(w, 200, result)
}

func (h *Handler) GetPembayaran(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	siswaID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "ID tidak valid")
		return
	}

	result, err := h.service.GetPembayaran(user.SekolahID, user.UserID, siswaID)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", "Data tidak ditemukan")
		return
	}
	if result == nil {
		result = []Pembayaran{}
	}
	response.JSON(w, 200, result)
}

func (h *Handler) CreatePembayaran(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	var req CreatePembayaranRequest
	if err := validator.DecodeJSON(r, &req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	allowedMetode := []string{"transfer", "cash"}
	errs := validator.Collect(
		validator.Required("tanggal", req.Tanggal),
		validator.Required("metode", req.Metode),
		validator.InList("metode", req.Metode, allowedMetode),
	)
	if req.TagihanID == 0 {
		errs = append(errs, validator.ValidationError{Field: "tagihan_id", Message: "wajib diisi"})
	}
	if req.Jumlah <= 0 {
		errs = append(errs, validator.ValidationError{Field: "jumlah", Message: "harus lebih dari 0"})
	}
	if req.Metode == "transfer" && req.RekeningSekolahID == 0 {
		errs = append(errs, validator.ValidationError{Field: "rekening_sekolah_id", Message: "wajib diisi untuk metode transfer"})
	}
	if len(errs) > 0 {
		response.ErrorWithDetails(w, 400, "VALIDATION_ERROR", "Data tidak valid", errs)
		return
	}

	id, err := h.service.CreatePembayaran(user.SekolahID, user.UserID, req)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", "Tagihan tidak ditemukan atau tidak memiliki akses")
		return
	}

	response.JSON(w, 201, map[string]int64{"id": id})
}

type GatewayPayRequest struct {
	Provider string `json:"provider"`
}

func (h *Handler) InitiateGatewayPayment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		response.Error(w, 401, "UNAUTHORIZED", "Belum login")
		return
	}

	tagihanID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "ID tagihan tidak valid")
		return
	}

	if h.gateway == nil {
		response.Error(w, 400, "GATEWAY_NOT_AVAILABLE", "Payment gateway tidak tersedia")
		return
	}

	var req GatewayPayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	if req.Provider != "midtrans" && req.Provider != "xendit" {
		response.Error(w, 400, "INVALID_PROVIDER", "Provider harus midtrans atau xendit")
		return
	}

	if !h.gateway.HasProvider(req.Provider) {
		response.Error(w, 400, "PROVIDER_NOT_CONFIGURED", "Payment provider belum dikonfigurasi")
		return
	}

	ok, siswaID, err := h.service.repo.TagihanBelongsToLinkedSiswa(user.SekolahID, user.UserID, tagihanID)
	if err != nil {
		if err == sql.ErrNoRows {
			response.Error(w, 404, "NOT_FOUND", "Tagihan tidak ditemukan")
			return
		}
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal memverifikasi akses")
		return
	}
	if !ok {
		response.Error(w, 404, "NOT_FOUND", "Tagihan tidak ditemukan atau tidak memiliki akses")
		return
	}
	_ = siswaID

	result, err := h.gateway.InitiateTransaction(user.SekolahID, tagihanID, user.UserID, req.Provider)
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "tagihan sudah lunas":
			response.Error(w, 409, "TAGIHAN_LUNAS", "Tagihan sudah lunas, tidak dapat membuat transaksi baru")
		case "payment provider not configured":
			response.Error(w, 400, "PROVIDER_NOT_CONFIGURED", "Payment provider belum dikonfigurasi")
		case "tagihan not found":
			response.Error(w, 404, "NOT_FOUND", "Tagihan tidak ditemukan")
		default:
			response.Error(w, 500, "INTERNAL_ERROR", fmt.Sprintf("Gagal membuat transaksi: %s", errMsg))
		}
		return
	}

	response.JSON(w, 200, result)
}

func (h *Handler) GatewayProviders(w http.ResponseWriter, r *http.Request) {
	if h.gateway == nil {
		response.JSON(w, 200, map[string]interface{}{"providers": []string{}})
		return
	}
	providers := h.gateway.AvailableProviders()
	if providers == nil {
		providers = []string{}
	}
	response.JSON(w, 200, map[string]interface{}{"providers": providers})
}

func (h *Handler) DashboardSiswa(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	result, err := h.service.DashboardSiswa(user.SekolahID, user.UserID)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data dashboard")
		return
	}
	response.JSON(w, 200, result)
}

func (h *Handler) DashboardOrangtua(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	result, err := h.service.DashboardOrangtua(user.SekolahID, user.UserID)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data dashboard")
		return
	}
	response.JSON(w, 200, result)
}

func (h *Handler) DashboardGuru(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	result, err := h.service.DashboardGuru(user.SekolahID, user.UserID)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data dashboard")
		return
	}
	response.JSON(w, 200, result)
}

func (h *Handler) ListGuruKelas(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	result, err := h.service.GetGuruKelas(user.SekolahID, user.UserID)
	if err != nil {
		response.Error(w, 500, "INTERNAL_ERROR", "Gagal mengambil data kelas")
		return
	}
	if result == nil {
		result = []GuruKelas{}
	}
	response.JSON(w, 200, result)
}

func (h *Handler) ListGuruSiswaByKelas(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	kelasID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "ID tidak valid")
		return
	}

	result, err := h.service.GetGuruSiswaByKelas(user.SekolahID, user.UserID, kelasID)
	if err != nil {
		response.Error(w, 404, "NOT_FOUND", "Kelas tidak ditemukan")
		return
	}
	if result == nil {
		result = []GuruSiswa{}
	}
	response.JSON(w, 200, result)
}
