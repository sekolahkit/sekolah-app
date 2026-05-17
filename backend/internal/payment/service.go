package payment

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Service struct {
	repo     *Repository
	gateways map[string]Gateway
}

func NewService(repo *Repository, gateways ...Gateway) *Service {
	m := make(map[string]Gateway, len(gateways))
	for _, g := range gateways {
		m[g.Provider()] = g
	}
	return &Service{repo: repo, gateways: m}
}

func (s *Service) AvailableProviders() []string {
	var providers []string
	for _, g := range s.gateways {
		providers = append(providers, g.Provider())
	}
	return providers
}

func (s *Service) HasProvider(provider string) bool {
	_, ok := s.gateways[provider]
	return ok
}

func (s *Service) InitiateTransaction(sekolahID, tagihanID, createdBy int64, provider string) (*GatewayTransaksi, error) {
	gw, ok := s.gateways[provider]
	if !ok {
		return nil, ErrProviderNotConfig
	}

	tagihan, err := s.repo.FindTagihanByIDAndSekolah(tagihanID, sekolahID)
	if err != nil {
		return nil, fmt.Errorf("find tagihan: %w", err)
	}
	if tagihan == nil {
		return nil, fmt.Errorf("tagihan not found")
	}

	if tagihan.Status == "lunas" {
		return nil, ErrTagihanLunas
	}

	existing, err := s.repo.FindPendingTransaction(tagihanID, provider)
	if err != nil {
		return nil, fmt.Errorf("check existing transaction: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	orderID := fmt.Sprintf("%d", tagihanID)

	gwResp, err := gw.CreateTransaction(context.Background(), CreateTransactionRequest{
		OrderID:    orderID,
		Amount:     tagihan.Nominal,
		Description: fmt.Sprintf("Pembayaran tagihan #%d", tagihanID),
		CustomerID:  fmt.Sprintf("%d", sekolahID),
	})
	if err != nil {
		return nil, fmt.Errorf("create gateway transaction: %w", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	transaksi := &GatewayTransaksi{
		SekolahID:       sekolahID,
		TagihanID:       tagihanID,
		Provider:        provider,
		OrderID:         orderID,
		PaymentGatewayID: gwResp.PaymentGatewayID,
		PaymentURL:      gwResp.PaymentURL,
		Amount:          tagihan.Nominal,
		Status:          "pending",
		ExpiresAt:       expiresAt.Format("2006-01-02 15:04:05"),
		CreatedBy:       createdBy,
	}

	id, err := s.repo.InsertGatewayTransaksi(transaksi)
	if err != nil {
		return nil, fmt.Errorf("save gateway transaction: %w", err)
	}

	transaksi.ID = id
	return transaksi, nil
}

func (s *Service) ProcessCallback(provider string, payload []byte, headers map[string]string) (*CallbackResult, error) {
	gw, ok := s.gateways[provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}

	result, err := gw.HandleCallback(payload, headers)
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.FindCallbackLog(provider, result.PaymentGatewayID)
	if err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}
	if existing != nil {
		return result, ErrDuplicateCallback
	}

	tagihan, err := s.repo.FindTagihanByOrderID(result.OrderID)
	if err != nil {
		return nil, fmt.Errorf("find tagihan: %w", err)
	}
	if tagihan == nil {
		return nil, fmt.Errorf("tagihan not found for order: %s", result.OrderID)
	}

	logEntry := &CallbackLog{
		Provider:         provider,
		PaymentGatewayID: result.PaymentGatewayID,
		OrderID:          result.OrderID,
		Status:           string(result.Status),
		Amount:           result.Amount,
		SekolahID:        tagihan.SekolahID,
		Processed:        false,
	}

	if result.Status != StatusSuccess {
		logEntry.Processed = true
		if err := s.repo.InsertCallbackLog(logEntry); err != nil {
			log.Printf("payment: failed to log callback: %v", err)
		}

		gwStatus := mapGatewayStatus(result.Status)
		if err := s.repo.UpdateGatewayTransaksiStatus(result.OrderID, gwStatus); err != nil {
			log.Printf("payment: failed to update gateway_transaksi status: %v", err)
		}

		return result, nil
	}

	paidAmount, err := s.repo.GetPaidAmount(tagihan.ID)
	if err != nil {
		return nil, fmt.Errorf("get paid amount: %w", err)
	}

	if paidAmount+result.Amount > tagihan.Nominal {
		logEntry.Processed = true
		if err := s.repo.InsertCallbackLog(logEntry); err != nil {
			log.Printf("payment: failed to log callback: %v", err)
		}
		return result, ErrOverpay
	}

	err = s.repo.InsertPembayaranAndUpdateTagihan(
		tagihan.ID, tagihan.SiswaID, tagihan.SekolahID,
		result.Amount, provider, result.PaymentGatewayID,
	)
	if err != nil {
		return nil, fmt.Errorf("process payment: %w", err)
	}

	if err := s.repo.UpdateGatewayTransaksiStatus(result.OrderID, "paid"); err != nil {
		log.Printf("payment: failed to update gateway_transaksi status: %v", err)
	}

	logEntry.Processed = true
	if err := s.repo.InsertCallbackLog(logEntry); err != nil {
		log.Printf("payment: failed to log callback: %v", err)
	}

	return result, nil
}

func mapGatewayStatus(s Status) string {
	switch s {
	case StatusSuccess:
		return "paid"
	case StatusFailed:
		return "failed"
	case StatusExpired:
		return "expired"
	default:
		return "pending"
	}
}
