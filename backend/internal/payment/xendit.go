package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type XenditConfig struct {
	SecretKey      string
	CallbackToken  string
}

type xenditGateway struct {
	cfg XenditConfig
}

func NewXendit(cfg XenditConfig) Gateway {
	return &xenditGateway{cfg: cfg}
}

func (x *xenditGateway) Provider() string {
	return "xendit"
}

func (x *xenditGateway) CreateTransaction(_ context.Context, req CreateTransactionRequest) (*CreateTransactionResponse, error) {
	return &CreateTransactionResponse{
		PaymentURL:       fmt.Sprintf("https://checkout-staging.xendit.co/web/%s", req.OrderID),
		PaymentGatewayID: req.OrderID,
		Provider:         "xendit",
	}, nil
}

func (x *xenditGateway) HandleCallback(payload []byte, headers map[string]string) (*CallbackResult, error) {
	token := headers["x-callback-token"]
	if token == "" {
		token = headers["X-Callback-Token"]
	}
	if token != x.cfg.CallbackToken {
		return nil, ErrInvalidToken
	}

	var notif xenditNotification
	if err := json.Unmarshal(payload, &notif); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	status := x.mapStatus(notif.Status)

	return &CallbackResult{
		OrderID:          notif.ExternalID,
		PaymentGatewayID: notif.ID,
		Status:           status,
		Amount:           notif.Amount,
		Provider:         "xendit",
	}, nil
}

func (x *xenditGateway) GetStatus(_ context.Context, orderID string) (*TransactionStatus, error) {
	return &TransactionStatus{
		OrderID:          orderID,
		PaymentGatewayID: orderID,
		Status:           StatusPending,
	}, nil
}

func (x *xenditGateway) mapStatus(status string) Status {
	switch status {
	case "PAID", "SETTLED":
		return StatusSuccess
	case "EXPIRED":
		return StatusExpired
	case "FAILED":
		return StatusFailed
	default:
		return StatusPending
	}
}

func (x *xenditGateway) ComputeHMAC(payload []byte) string {
	mac := hmac.New(sha256.New, []byte(x.cfg.SecretKey))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

type xenditNotification struct {
	ID         string `json:"id"`
	ExternalID string `json:"external_id"`
	Status     string `json:"status"`
	Amount     int64  `json:"amount"`
}
