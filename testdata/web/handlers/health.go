package handlers

import (
	"encoding/json"
	"net/http"
)

type HealthStatus struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func HandleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		status := HealthStatus{
			Status:  "ok",
			Version: "1.0.0",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}
}
