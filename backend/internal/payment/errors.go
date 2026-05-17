package payment

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrInvalidSignature  = errors.New("invalid signature")
	ErrInvalidToken      = errors.New("invalid callback token")
	ErrDuplicateCallback = errors.New("duplicate callback")
	ErrOverpay           = errors.New("payment amount exceeds tagihan")
	ErrTagihanLunas      = errors.New("tagihan sudah lunas")
	ErrProviderNotConfig = errors.New("payment provider not configured")
)

func parseAmount(s string) int64 {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, ".00", "", 1)
	s = strings.Replace(s, ",", "", -1)
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}
