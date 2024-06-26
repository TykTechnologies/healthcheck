package main

import (
	"fmt"
	hc "github.com/TykTechnologies/healthcheck"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	// Initialize HealthChecker
	myHealthChecker := hc.NewHealthChecker()

	// Register a liveness check
	myHealthChecker.RegisterCheck("PingCheck", hc.Required, func() (hc.HealthStatus, error) {
		return hc.StatusPass, nil
	})

	myHealthChecker.RegisterCheck("cachedCheck", hc.Required, func() (hc.HealthStatus, error) {
		status := randomStatusCheck() // Use this in a real registration call
		return status(), nil
	}).WithCache(3)

	// Create a DatabaseCheck with a ticker that fires every 30 seconds
	dbCheck := NewDatabaseCheck(30 * time.Second)
	// Register the cached readiness check
	myHealthChecker.RegisterCheck("Database", hc.Required, func() (hc.HealthStatus, error) {
		result := dbCheck.GetCachedResult()
		return result.Status, nil
	})

	// add handlers
	fmt.Printf("Visit for detailed healthcheck status: http://localhost:9000//health/healthcheck")
	http.HandleFunc("/health/healthcheck", myHealthChecker.HTTPHandler())

	http.ListenAndServe(":9000", nil)
}

func randomStatusCheck() func() hc.HealthStatus {
	// Seed the random number generator (important to do this once, not in the returned function)
	rand.Seed(time.Now().UnixNano())

	return func() hc.HealthStatus {
		if rand.Intn(2) == 0 { // Randomly returns 0 or 1
			return hc.StatusPass
		} else {
			return hc.StatusFail
		}
	}
}
