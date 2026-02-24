package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aneeshchawla/kubetools/backend/internal/models"
)

func WriteSuccess(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(models.APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().UTC(),
	})
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    code,
			Message: message,
		},
		Timestamp: time.Now().UTC(),
	})
}
