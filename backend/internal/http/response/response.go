package response

import (
	"encoding/json"
	"net/http"
)

type ErrorPayload struct {
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}

func Error(w http.ResponseWriter, status int, message string, fields map[string]string) {
	JSON(w, status, ErrorPayload{
		Message: message,
		Fields:  fields,
	})
}
