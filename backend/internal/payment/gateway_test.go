package payment

import (
	"context"
	"errors"
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

func TestMockGatewayCharge_Declined(t *testing.T) {
	gateway := NewMockGateway()

	resp, err := gateway.Charge(context.Background(), ChargeRequest{
		Amount:        19.99,
		PaymentMethod: "card_decline",
		HoldID:        "hold_1",
		UserID:        "usr_1",
	})
	if !errors.Is(err, ErrPaymentDeclined) {
		t.Fatalf("expected ErrPaymentDeclined, got %v", err)
	}
	if resp.Status != "FAILED" {
		t.Fatalf("expected FAILED status, got %s", resp.Status)
	}
	if resp.Reason == "" {
		t.Fatal("expected failure reason for declined payment")
	}
}

func TestMockGatewayCharge_Timeout(t *testing.T) {
	gateway := NewMockGateway()

	resp, err := gateway.Charge(context.Background(), ChargeRequest{
		Amount:        19.99,
		PaymentMethod: "card_timeout",
		HoldID:        "hold_1",
		UserID:        "usr_1",
	})
	if !errors.Is(err, ErrGatewayTimeout) {
		t.Fatalf("expected ErrGatewayTimeout, got %v", err)
	}
	if resp.GatewayTxnID != "" || resp.Status != "" || resp.Reason != "" {
		t.Fatalf("expected empty response on timeout, got %+v", resp)
	}
}
