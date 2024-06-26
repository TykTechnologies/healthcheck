package main

import (
	hc "github.com/TykTechnologies/healthcheck"
	"sync"
	"time"
)

// DatabaseCheck simulates a database check with caching and a ticker to perform underground update of the status
type DatabaseCheck struct {
	mu           sync.Mutex
	cachedResult hc.CheckResult
	ticker       *time.Ticker
}

func NewDatabaseCheck(updateInterval time.Duration) *DatabaseCheck {
	dbCheck := &DatabaseCheck{
		cachedResult: hc.CheckResult{Name: "Database", Status: hc.StatusFail}, // default status
		ticker:       time.NewTicker(updateInterval),
	}

	// Start the periodic update
	go func() {
		for {
			select {
			case <-dbCheck.ticker.C:
				dbCheck.updateCache()
			}
		}
	}()

	// Perform the initial update immediately
	dbCheck.updateCache()

	return dbCheck
}

func (db *DatabaseCheck) updateCache() {
	// Perform the check (simulated here)
	time.Sleep(100 * time.Millisecond) // Simulate time-consuming work

	db.mu.Lock()
	defer db.mu.Unlock()

	// Update the cached result
	// Logic here will be replaced with actual check logic
	if db.cachedResult.Status == hc.StatusPass {
		db.cachedResult.Status = hc.StatusFail
	} else {
		db.cachedResult.Status = hc.StatusPass
	}

	db.cachedResult.ObservationTS = time.Now()
}

func (db *DatabaseCheck) GetCachedResult() hc.CheckResult {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.cachedResult
}
