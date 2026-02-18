package observability

import (
	"encoding/json"
	"net/http"
	"sync"
)

// HealthChecker provides health status.
type HealthChecker struct {
	checks map[string]string
	mu     sync.RWMutex
}

// NewHealthChecker creates a health checker.
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]string),
	}
}

// Set sets a check status.
func (h *HealthChecker) Set(name, status string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = status
}

// Handler returns an HTTP handler for GET /health.
func (h *HealthChecker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		h.mu.RLock()
		checks := make(map[string]string, len(h.checks))
		for k, v := range h.checks {
			checks[k] = v
		}
		h.mu.RUnlock()

		status := "ok"
		for _, v := range checks {
			if v != "ok" {
				status = "degraded"
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": status,
			"checks": checks,
		})
	}
}
