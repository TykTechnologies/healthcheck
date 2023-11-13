package TykHealthcheck

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make([]*Check, 0),
		mu:     sync.RWMutex{},
	}
}

func (hc *HealthChecker) RegisterCheck(name string, importance CheckImportance, checkFunc CheckFunc) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	wrappedCheckFunc := func() CheckResult {
		startTime := time.Now()
		status, err := checkFunc()
		duration := time.Since(startTime)

		var output string
		if err != nil {
			output = err.Error()
		}

		return CheckResult{
			Name:          name,
			Status:        status,
			ObservationTS: time.Now().Add(-duration),
			Output:        output,
		}
	}

	check := &Check{
		Name:       name,
		Importance: importance,
		Perform:    wrappedCheckFunc,
	}

	hc.checks = append(hc.checks, check)
}

func (hc *HealthChecker) PerformChecks() HealthCheckResponse {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	var results []CheckResult
	var overallStatus HealthStatus = StatusPass
	var statusCode int = 200

	for _, check := range hc.checks {
		result := check.Perform()
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
		OverallStatus: string(overallStatus),
		StatusCode:    statusCode,
		Components:    results,
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
