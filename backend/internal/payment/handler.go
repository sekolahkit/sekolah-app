package payment

import (
	"io"
	"net/http"

	"github.com/Sekolahkit/sekolah-app/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
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
