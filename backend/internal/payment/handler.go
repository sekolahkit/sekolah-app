package payment

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/Sekolahkit/sekolah-app/pkg/middleware"
	"github.com/Sekolahkit/sekolah-app/pkg/response"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type InitiateRequest struct {
	Provider string `json:"provider"`
}

type InitiateResponse struct {
	Provider        string `json:"provider"`
	OrderID         string `json:"order_id"`
	PaymentURL      string `json:"payment_url"`
	PaymentGatewayID string `json:"payment_gateway_id"`
	Status          string `json:"status"`
}

func (h *Handler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
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

	var req InitiateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, 400, "INVALID_REQUEST", "Request body tidak valid")
		return
	}

	if req.Provider != "midtrans" && req.Provider != "xendit" {
		response.Error(w, 400, "INVALID_PROVIDER", "Provider harus midtrans atau xendit")
		return
	}

	if !h.service.HasProvider(req.Provider) {
		response.Error(w, 400, "PROVIDER_NOT_CONFIGURED", "Payment provider belum dikonfigurasi")
		return
	}

	result, err := h.service.InitiateTransaction(user.SekolahID, tagihanID, user.UserID, req.Provider)
	if err != nil {
		switch err {
		case ErrTagihanLunas:
			response.Error(w, 409, "TAGIHAN_LUNAS", "Tagihan sudah lunas, tidak dapat membuat transaksi baru")
		case ErrProviderNotConfig:
			response.Error(w, 400, "PROVIDER_NOT_CONFIGURED", "Payment provider belum dikonfigurasi")
		default:
			response.Error(w, 404, "NOT_FOUND", "Tagihan tidak ditemukan atau tidak memiliki akses")
		}
		return
	}

	resp := InitiateResponse{
		Provider:        result.Provider,
		OrderID:         result.OrderID,
		PaymentURL:      result.PaymentURL,
		PaymentGatewayID: result.PaymentGatewayID,
		Status:          result.Status,
	}

	response.JSON(w, 200, resp)
}

func (h *Handler) ListProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.service.AvailableProviders()
	if providers == nil {
		providers = []string{}
	}
	response.JSON(w, 200, map[string]interface{}{
		"providers": providers,
	})
}

func (h *Handler) MidtransCallback(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, 400, "INVALID_BODY", "Gagal membaca request body")
		return
	}
	defer r.Body.Close()

	result, err := h.service.ProcessCallback("midtrans", body, extractHeaders(r))
	if err != nil {
		h.handleCallbackError(w, err, result)
		return
	}

	response.JSON(w, 200, map[string]string{"status": "ok"})
}

func (h *Handler) XenditCallback(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, 400, "INVALID_BODY", "Gagal membaca request body")
		return
	}
	defer r.Body.Close()

	result, err := h.service.ProcessCallback("xendit", body, extractHeaders(r))
	if err != nil {
		h.handleCallbackError(w, err, result)
		return
	}

	response.JSON(w, 200, map[string]string{"status": "ok"})
}

func (h *Handler) handleCallbackError(w http.ResponseWriter, err error, _ *CallbackResult) {
	switch err {
	case ErrInvalidSignature, ErrInvalidToken:
		response.Error(w, 401, "INVALID_SIGNATURE", "Signature/token tidak valid")
	case ErrDuplicateCallback:
		response.JSON(w, 200, map[string]string{"status": "duplicate"})
	case ErrOverpay:
		response.Error(w, 422, "OVERPAY", "Pembayaran melebihi nominal tagihan")
	default:
		response.Error(w, 500, "CALLBACK_ERROR", err.Error())
	}
}

func extractHeaders(r *http.Request) map[string]string {
	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}
	return headers
}
