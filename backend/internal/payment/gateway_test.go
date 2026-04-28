package payment

import (
	"context"
	"strings"
	"testing"
)

func TestNewMockGateway(t *testing.T) {
	gateway := NewMockGateway()
	if gateway == nil {
		t.Fatal("expected non-nil mock gateway")
	}
}

func TestMockGatewayCharge_Success(t *testing.T) {
	gateway := NewMockGateway()

	resp, err := gateway.Charge(context.Background(), ChargeRequest{
		Amount:        19.99,
		PaymentMethod: "card",
		CardNumber:    "4111111111111111",
		CardExpiry:    "12/29",
		CardCVV:       "123",
		HoldID:        "hold_1",
		UserID:        "usr_1",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp.Status != "SUCCESS" {
		t.Fatalf("expected SUCCESS, got %s", resp.Status)
	}
	if !strings.HasPrefix(resp.GatewayTxnID, "gw_txn_") {
		t.Fatalf("expected gateway transaction id prefix gw_txn_, got %s", resp.GatewayTxnID)
	}
}

func TestMockGatewayCharge_AlwaysSuccessForArbitraryInput(t *testing.T) {
	gateway := NewMockGateway()

	resp, err := gateway.Charge(context.Background(), ChargeRequest{
		Amount:        19.99,
		PaymentMethod: "some-random-value",
		CardNumber:    "not-a-real-card",
		CardExpiry:    "00/00",
		CardCVV:       "000",
		HoldID:        "hold_1",
		UserID:        "usr_1",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp.Status != "SUCCESS" {
		t.Fatalf("expected SUCCESS, got %s", resp.Status)
	}
	if !strings.HasPrefix(resp.GatewayTxnID, "gw_txn_") {
		t.Fatalf("expected gateway transaction id prefix gw_txn_, got %s", resp.GatewayTxnID)
	}
}
