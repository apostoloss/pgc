package health

import (
	"encoding/json"
	"my-solution/internal/version"
	"net/http"
)

// HealthResponse represents the JSON structure returned by the health endpoint.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// Handler is the HTTP handler for the /healthz endpoint.

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HealthResponse{
		Status:  "ok",
		Version: version.Get(), // From shared internal/version
	})
}
