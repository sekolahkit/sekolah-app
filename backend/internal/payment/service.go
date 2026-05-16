package payment

import (
	"fmt"
	"log"
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

	logEntry.Processed = true
	if err := s.repo.InsertCallbackLog(logEntry); err != nil {
		log.Printf("payment: failed to log callback: %v", err)
	}

	return result, nil
}
