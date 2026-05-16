package payment

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type MidtransConfig struct {
	ServerKey string
	ClientKey string
	IsProduction bool
}

type midtransGateway struct {
	cfg MidtransConfig
}

func NewMidtrans(cfg MidtransConfig) Gateway {
	return &midtransGateway{cfg: cfg}
}

func (m *midtransGateway) Provider() string {
	return "midtrans"
}

func (m *midtransGateway) CreateTransaction(_ context.Context, req CreateTransactionRequest) (*CreateTransactionResponse, error) {
	return &CreateTransactionResponse{
		PaymentURL:       fmt.Sprintf("https://app.sandbox.midtrans.com/snap/v2/vtweb/%s", req.OrderID),
		PaymentGatewayID: req.OrderID,
		Provider:         "midtrans",
	}, nil
}

func (m *midtransGateway) HandleCallback(payload []byte, _ map[string]string) (*CallbackResult, error) {
	var notif midtransNotification
	if err := json.Unmarshal(payload, &notif); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	expectedSig := m.computeSignature(notif.OrderID, notif.StatusCode, notif.GrossAmount)
	if !strings.EqualFold(notif.SignatureKey, expectedSig) {
		return nil, ErrInvalidSignature
	}

	status := m.mapStatus(notif.TransactionStatus, notif.FraudStatus)

	return &CallbackResult{
		OrderID:          notif.OrderID,
		PaymentGatewayID: notif.TransactionID,
		Status:           status,
		Amount:           parseAmount(notif.GrossAmount),
		Provider:         "midtrans",
	}, nil
}

func (m *midtransGateway) GetStatus(_ context.Context, orderID string) (*TransactionStatus, error) {
	return &TransactionStatus{
		OrderID:          orderID,
		PaymentGatewayID: orderID,
		Status:           StatusPending,
	}, nil
}

func (m *midtransGateway) computeSignature(orderID, statusCode, grossAmount string) string {
	raw := orderID + statusCode + grossAmount + m.cfg.ServerKey
	h := sha512.New()
	h.Write([]byte(raw))
	return hex.EncodeToString(h.Sum(nil))
}

func (m *midtransGateway) mapStatus(transactionStatus, fraudStatus string) Status {
	switch transactionStatus {
	case "capture":
		if fraudStatus == "accept" {
			return StatusSuccess
		}
		return StatusPending
	case "settlement":
		return StatusSuccess
	case "deny", "cancel":
		return StatusFailed
	case "expire":
		return StatusExpired
	case "pending":
		return StatusPending
	default:
		return StatusPending
	}
}

type midtransNotification struct {
	TransactionID     string `json:"transaction_id"`
	OrderID           string `json:"order_id"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey       string `json:"signature_key"`
}
