package api

import (
	"encoding/json"
	"net/http"
)

func WriteJson(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "Failed to encode product to JSON", http.StatusInternalServerError)
	}
}

func WriteContent(w http.ResponseWriter, payload []byte, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}
