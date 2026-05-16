package payment

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"testing"
)

func TestMidtransInvalidSignatureRejected(t *testing.T) {
	gw := NewMidtrans(MidtransConfig{ServerKey: "test-server-key"})

	notif := midtransNotification{
		TransactionID:     "txn-001",
		OrderID:           "order-001",
		TransactionStatus: "settlement",
		FraudStatus:       "accept",
		StatusCode:        "200",
		GrossAmount:       "100000.00",
		SignatureKey:       "invalid-signature",
	}
	payload, _ := json.Marshal(notif)

	_, err := gw.HandleCallback(payload, nil)
	if err != ErrInvalidSignature {
		t.Fatalf("expected ErrInvalidSignature, got: %v", err)
	}
}

func TestMidtransValidSignatureAccepted(t *testing.T) {
	serverKey := "test-server-key"
	gw := NewMidtrans(MidtransConfig{ServerKey: serverKey})

	orderID := "order-001"
	statusCode := "200"
	grossAmount := "100000.00"
	raw := orderID + statusCode + grossAmount + serverKey
	h := sha512.New()
	h.Write([]byte(raw))
	sig := hex.EncodeToString(h.Sum(nil))

	notif := midtransNotification{
		TransactionID:     "txn-001",
		OrderID:           orderID,
		TransactionStatus: "settlement",
		FraudStatus:       "accept",
		StatusCode:        statusCode,
		GrossAmount:       grossAmount,
		SignatureKey:       sig,
	}
	payload, _ := json.Marshal(notif)

	result, err := gw.HandleCallback(payload, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusSuccess {
		t.Fatalf("expected StatusSuccess, got: %s", result.Status)
	}
	if result.OrderID != orderID {
		t.Fatalf("expected order %s, got: %s", orderID, result.OrderID)
	}
	if result.Amount != 100000 {
		t.Fatalf("expected amount 100000, got: %d", result.Amount)
	}
}

func TestMidtransStatusMapping(t *testing.T) {
	serverKey := "key"
	gw := NewMidtrans(MidtransConfig{ServerKey: serverKey})

	tests := []struct {
		txStatus    string
		fraudStatus string
		expected    Status
	}{
		{"settlement", "", StatusSuccess},
		{"capture", "accept", StatusSuccess},
		{"capture", "challenge", StatusPending},
		{"pending", "", StatusPending},
		{"deny", "", StatusFailed},
		{"cancel", "", StatusFailed},
		{"expire", "", StatusExpired},
	}

	for _, tc := range tests {
		orderID := "o1"
		statusCode := "200"
		grossAmount := "1000"
		raw := orderID + statusCode + grossAmount + serverKey
		h := sha512.New()
		h.Write([]byte(raw))
		sig := hex.EncodeToString(h.Sum(nil))

		notif := midtransNotification{
			TransactionID:     "t1",
			OrderID:           orderID,
			TransactionStatus: tc.txStatus,
			FraudStatus:       tc.fraudStatus,
			StatusCode:        statusCode,
			GrossAmount:       grossAmount,
			SignatureKey:       sig,
		}
		payload, _ := json.Marshal(notif)

		result, err := gw.HandleCallback(payload, nil)
		if err != nil {
			t.Fatalf("status=%s fraud=%s: unexpected error: %v", tc.txStatus, tc.fraudStatus, err)
		}
		if result.Status != tc.expected {
			t.Errorf("status=%s fraud=%s: expected %s, got %s", tc.txStatus, tc.fraudStatus, tc.expected, result.Status)
		}
	}
}

func TestXenditInvalidTokenRejected(t *testing.T) {
	gw := NewXendit(XenditConfig{
		SecretKey:     "secret",
		CallbackToken: "valid-token",
	})

	notif := xenditNotification{
		ID:         "xnd-001",
		ExternalID: "order-001",
		Status:     "PAID",
		Amount:     50000,
	}
	payload, _ := json.Marshal(notif)

	_, err := gw.HandleCallback(payload, map[string]string{
		"X-Callback-Token": "wrong-token",
	})
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got: %v", err)
	}
}

func TestXenditMissingTokenRejected(t *testing.T) {
	gw := NewXendit(XenditConfig{
		SecretKey:     "secret",
		CallbackToken: "valid-token",
	})

	notif := xenditNotification{
		ID:         "xnd-001",
		ExternalID: "order-001",
		Status:     "PAID",
		Amount:     50000,
	}
	payload, _ := json.Marshal(notif)

	_, err := gw.HandleCallback(payload, map[string]string{})
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got: %v", err)
	}
}

func TestXenditValidTokenAccepted(t *testing.T) {
	token := "valid-token"
	gw := NewXendit(XenditConfig{
		SecretKey:     "secret",
		CallbackToken: token,
	})

	notif := xenditNotification{
		ID:         "xnd-001",
		ExternalID: "order-001",
		Status:     "PAID",
		Amount:     50000,
	}
	payload, _ := json.Marshal(notif)

	result, err := gw.HandleCallback(payload, map[string]string{
		"X-Callback-Token": token,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusSuccess {
		t.Fatalf("expected StatusSuccess, got: %s", result.Status)
	}
	if result.Amount != 50000 {
		t.Fatalf("expected amount 50000, got: %d", result.Amount)
	}
}

func TestXenditStatusMapping(t *testing.T) {
	token := "tok"
	gw := NewXendit(XenditConfig{SecretKey: "s", CallbackToken: token})

	tests := []struct {
		status   string
		expected Status
	}{
		{"PAID", StatusSuccess},
		{"SETTLED", StatusSuccess},
		{"EXPIRED", StatusExpired},
		{"FAILED", StatusFailed},
		{"PENDING", StatusPending},
	}

	for _, tc := range tests {
		notif := xenditNotification{ID: "x1", ExternalID: "o1", Status: tc.status, Amount: 1000}
		payload, _ := json.Marshal(notif)
		result, err := gw.HandleCallback(payload, map[string]string{"X-Callback-Token": token})
		if err != nil {
			t.Fatalf("status=%s: unexpected error: %v", tc.status, err)
		}
		if result.Status != tc.expected {
			t.Errorf("status=%s: expected %s, got %s", tc.status, tc.expected, result.Status)
		}
	}
}

func TestParseAmount(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"100000.00", 100000},
		{"100000", 100000},
		{"50000.00", 50000},
		{"0", 0},
	}
	for _, tc := range tests {
		got := parseAmount(tc.input)
		if got != tc.expected {
			t.Errorf("parseAmount(%q) = %d, want %d", tc.input, got, tc.expected)
		}
	}
}
