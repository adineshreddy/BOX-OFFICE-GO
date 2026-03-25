package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON_WritesStatusAndBody(t *testing.T) {
	rec := httptest.NewRecorder()
	JSON(rec, http.StatusCreated, map[string]string{"ok": "yes"})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected application/json, got %s", rec.Header().Get("Content-Type"))
	}

	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json body: %v", err)
	}
	if payload["ok"] != "yes" {
		t.Fatalf("unexpected payload: %v", payload)
	}
}

func TestError_WritesErrorPayload(t *testing.T) {
	rec := httptest.NewRecorder()
	Error(rec, http.StatusBadRequest, "validation failed", map[string]string{"email": "required"})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var payload ErrorPayload
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json body: %v", err)
	}
	if payload.Message != "validation failed" || payload.Fields["email"] != "required" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}
