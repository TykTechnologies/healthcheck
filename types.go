package TykHealthcheck

import (
	"sync"
	"time"
)

// CheckFunc defines the signature for health check functions.
// Each health check function must conform to this signature to be registered and executed by the HealthChecker.
// The return values are as follows:
//   - HealthStatus: The status of the check, represented by predefined constants (StatusPass, StatusWarn, StatusFail).
//     This indicates the health of the component being checked.
//   - error: An error value that should be nil if the check was successful, or provide error details if the check failed.
type CheckFunc func() (HealthStatus, error)

type HealthStatus string

const (
	StatusPass HealthStatus = "pass"
	StatusWarn HealthStatus = "warn"
	StatusFail HealthStatus = "fail"
)

type CheckImportance string

const (
	Required CheckImportance = "required"
	Optional CheckImportance = "optional"
	Info     CheckImportance = "info"
)

// HealthChecker manages a set of checks
type HealthChecker struct {
	checks []*Check
	mu     sync.RWMutex
}

// Check represents an individual health check
type Check struct {
	Name       string
	Importance CheckImportance
	Perform    func() CheckResult
}

type HealthCheckResponse struct {
	OverallStatus string        `json:"overallStatus"`
	StatusCode    int           `json:"statusCode"`
	Components    []CheckResult `json:"components"`
}

// CheckResult represents the result of a health check
type CheckResult struct {
	Name          string       `json:"name"`
	Type          string       `json:"type"`
	Status        HealthStatus `json:"status"`
	ObservationTS time.Time    `json:"observationTS"`
	Output        string       `json:"output"`
}
