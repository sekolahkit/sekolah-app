package payment

import "context"

type Status string

const (
	StatusPending Status = "pending"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
	StatusExpired Status = "expired"
)

type CreateTransactionRequest struct {
	OrderID     string
	Amount      int64
	Description string
	CustomerID  string
}

type CreateTransactionResponse struct {
	PaymentURL       string
	PaymentGatewayID string
	Provider         string
}

type CallbackResult struct {
	OrderID          string
	PaymentGatewayID string
	Status           Status
	Amount           int64
	Provider         string
}

type TransactionStatus struct {
	OrderID          string
	PaymentGatewayID string
	Status           Status
	Amount           int64
}

type Gateway interface {
	Provider() string
	CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*CreateTransactionResponse, error)
	HandleCallback(payload []byte, headers map[string]string) (*CallbackResult, error)
	GetStatus(ctx context.Context, orderID string) (*TransactionStatus, error)
}
