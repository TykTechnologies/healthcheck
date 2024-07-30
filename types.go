package healthcheck

import (
	"time"

	"github.com/patrickmn/go-cache"
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
	// Required should be used when a check should be always ok
	Required CheckImportance = "required"
	// Optional can be used when a check is not required to be everytime ok, but if it fails then it not make the application fail
	Optional CheckImportance = "optional"
	// Info should be used when the check doesn't directly impact the operational status
	Info CheckImportance = "info"
)

type caching struct {
	cacheTTL int
	ticker   *time.Ticker
	stopChan chan struct{}

	cache *cache.Cache
}
