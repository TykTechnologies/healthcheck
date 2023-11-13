package main

import (
	hc "github.com/TykTechnologies/TykHealthcheck"
	"net/http"
	"time"
)

func main() {
	// Initialize HealthChecker
	readinessHealthChecker := hc.NewHealthChecker()
	livenessHealthChecker := hc.NewHealthChecker()

	// Register a liveness check
	livenessHealthChecker.RegisterCheck("PingCheck", hc.Required, func() (hc.HealthStatus, error) {
		return hc.StatusPass, nil
	})

	// Create a DatabaseCheck with a ticker that fires every 30 seconds
	dbCheck := NewDatabaseCheck(30 * time.Second)
	// Register the cached readiness check
	readinessHealthChecker.RegisterCheck("Database", hc.Required, func() (hc.HealthStatus, error) {
		result := dbCheck.GetCachedResult()

		return result.Status, nil
	})

	// Register a readiness check
	readinessHealthChecker.RegisterCheck("custom-check", hc.Required, func() (hc.HealthStatus, error) {
		time.Sleep(100 * time.Millisecond)
		return hc.StatusPass, nil
	})

	// add handlers
	http.HandleFunc("/health/live", livenessHealthChecker.HTTPHandler())
	http.HandleFunc("/health/ready", readinessHealthChecker.HTTPHandler())

	http.ListenAndServe(":9000", nil)
}
