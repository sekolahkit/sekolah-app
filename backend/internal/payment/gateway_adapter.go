package payment

import "github.com/Sekolahkit/sekolah-app/internal/selfservice"

type GatewayAdapter struct {
	service *Service
}

func NewGatewayAdapter(service *Service) *GatewayAdapter {
	return &GatewayAdapter{service: service}
}

var _ selfservice.GatewayPaymentInitiator = (*GatewayAdapter)(nil)

func (a *GatewayAdapter) InitiateTransaction(sekolahID, tagihanID, createdBy int64, provider string) (*selfservice.GatewayPaymentResult, error) {
	result, err := a.service.InitiateTransaction(sekolahID, tagihanID, createdBy, provider)
	if err != nil {
		return nil, err
	}
	return &selfservice.GatewayPaymentResult{
		Provider:        result.Provider,
		OrderID:         result.OrderID,
		PaymentURL:      result.PaymentURL,
		PaymentGatewayID: result.PaymentGatewayID,
		Status:          result.Status,
	}, nil
}

func (a *GatewayAdapter) HasProvider(provider string) bool {
	return a.service.HasProvider(provider)
}

func (a *GatewayAdapter) AvailableProviders() []string {
	return a.service.AvailableProviders()
}
