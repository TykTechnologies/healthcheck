package healthcheck

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// HealthChecker manages a set of checks
type HealthChecker struct {
	checks []*Check
	mu     sync.RWMutex
}

type HealthCheckResponse struct {
	Status     HealthStatus  `json:"status"`
	StatusCode int           `json:"status_code"`
	Components []CheckResult `json:"components"`
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make([]*Check, 0),
		mu:     sync.RWMutex{},
	}
}

func (hc *HealthChecker) RegisterCheck(name string, importance CheckImportance, checkFunc CheckFunc) *Check {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	check := &Check{
		Name:       name,
		Importance: importance,
	}
	// wrapp checkFunc to set the observationTS
	check.Perform = check.timedCheck(checkFunc)

	hc.checks = append(hc.checks, check)
	return check
}

func (hc *HealthChecker) PerformChecks() HealthCheckResponse {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	var results []CheckResult
	var overallStatus = StatusPass
	var statusCode = 200

	for _, check := range hc.checks {
		result := check.getStatus()
		results = append(results, result)

		if result.Status == StatusFail && check.Importance == Required {
			overallStatus = StatusFail
			statusCode = 500
		} else if result.Status == StatusWarn && overallStatus != StatusFail {
			overallStatus = StatusWarn
			statusCode = 429
		}
	}

	return HealthCheckResponse{
		Status:     overallStatus,
		StatusCode: statusCode,
		Components: results,
	}
}

func (hc *HealthChecker) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checkResults := hc.PerformChecks()
		serializedResults, err := json.Marshal(checkResults)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error serializing check results: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(checkResults.StatusCode)
		w.Write(serializedResults)
	}
}
