package healthcheck

import (
	"time"

	"github.com/patrickmn/go-cache"
)

// Check represents an individual health check
type Check struct {
	Name       string
	Importance CheckImportance
	Perform    func() CheckResult
	caching    *caching
}

// CheckResult represents the result of a health check
type CheckResult struct {
	Name          string       `json:"name"`
	Status        HealthStatus `json:"status"`
	ObservationTS time.Time    `json:"observation_ts"`
}

func (c *Check) WithCache(ttl int) {
	ttlDuration := time.Duration(ttl) * time.Second
	c.caching = &caching{
		cacheTTL: ttl,
		ticker:   time.NewTicker(ttlDuration),
		stopChan: make(chan struct{}),
		cache:    cache.New(ttlDuration, ttlDuration),
	}

	// Start the periodic update
	go func() {
		for {
			select {
			case <-c.caching.ticker.C:
				c.UpdateCache()
			case <-c.caching.stopChan:
				c.caching.ticker.Stop()
				return
			}
		}
	}()
}

func (c *Check) StopCacheTicker() {
	if c.caching != nil && c.caching.stopChan != nil {
		close(c.caching.stopChan) // Close the channel to signal the goroutine to stop
	}
}

func (c *Check) isCached() bool {
	return c.caching != nil
}

func (c *Check) getStatus() CheckResult {
	if c.isCached() {
		if cachedValue, found := c.caching.cache.Get(c.Name); found {
			return cachedValue.(CheckResult)
		}

		// If the cache was supposed to have the value but didn't, update the cache.
		c.UpdateCache()
		if cachedValue, found := c.caching.cache.Get(c.Name); found {
			return cachedValue.(CheckResult)
		}
	}

	return c.Perform()
}

func (c *Check) UpdateCache() {
	if !c.isCached() {
		return
	}

	res := c.Perform()
	c.caching.cache.Set(c.Name, res, cache.DefaultExpiration)
}

// set the observation time
func (c *Check) timedCheck(checkFunc CheckFunc) func() CheckResult {
	return func() CheckResult {
		observationTS := time.Now()

		status, _ := checkFunc()
		return CheckResult{
			Name:          c.Name,
			Status:        status,
			ObservationTS: observationTS,
		}
	}
}
