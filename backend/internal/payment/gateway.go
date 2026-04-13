package payment

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrPaymentDeclined = errors.New("payment declined")
	ErrGatewayTimeout  = errors.New("payment gateway timeout")
)

// ChargeRequest represents a request to charge a customer.
type ChargeRequest struct {
	Amount        float64
	PaymentMethod string
	HoldID        string
	UserID        string
}

// ChargeResponse represents the gateway's response to a charge request.
type ChargeResponse struct {
	GatewayTxnID string
	Status       string // "SUCCESS" or "FAILED"
	Reason       string // populated on failure
}

// Gateway abstracts a payment provider. Swap MockGateway for Stripe/Braintree in production.
type Gateway interface {
	Charge(ctx context.Context, req ChargeRequest) (ChargeResponse, error)
}

// MockGateway simulates payment processing for development/testing.
// It uses deterministic rules based on PaymentMethod to trigger different outcomes:
//
//	"card_decline" → returns ErrPaymentDeclined
//	"card_timeout" → returns ErrGatewayTimeout
//	any other value → succeeds
type MockGateway struct{}

func NewMockGateway() *MockGateway {
	return &MockGateway{}
}

func (g *MockGateway) Charge(_ context.Context, req ChargeRequest) (ChargeResponse, error) {
	switch req.PaymentMethod {
	case "card_decline":
		return ChargeResponse{
			Status: "FAILED",
			Reason: "card declined by issuing bank",
		}, ErrPaymentDeclined

	case "card_timeout":
		return ChargeResponse{}, ErrGatewayTimeout

	default:
		return ChargeResponse{
			GatewayTxnID: fmt.Sprintf("gw_txn_%d", time.Now().UnixNano()),
			Status:       "SUCCESS",
		}, nil
	}
}
