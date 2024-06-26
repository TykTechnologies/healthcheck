package healthcheck

import (
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestIsCached(t *testing.T) {
	testCases := []struct {
		name  string
		want  bool
		check *Check
	}{
		{
			name: "not cached",
			want: false,
			check: &Check{
				caching: nil,
			},
		},
		{
			name: "is cached",
			want: true,
			check: &Check{
				caching: &caching{
					cacheTTL: 10,
					cache:    cache.New(10, 10),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.check.isCached())
		})
	}
}

// TestTimedCheck tests the timedCheck function to ensure it returns a valid CheckResult.
func TestTimedCheck(t *testing.T) {
	// Mock CheckFunc that returns a predetermined status.
	mockCheckFunc := func() (HealthStatus, error) {
		return StatusPass, nil
	}

	testName := "TestCheck"
	c := Check{
		Name: testName,
	}

	resultFunc := c.timedCheck(mockCheckFunc)
	result := resultFunc()

	// Check the name
	if result.Name != testName {
		t.Errorf("Expected name %v, got %v", testName, result.Name)
	}

	// Check the observation timestamp
	if time.Since(result.ObservationTS) > time.Second {
		t.Errorf("Expected ObservationTS to be within the last second")
	}
}

func TestWithCache(t *testing.T) {
	status := StatusPass
	check := Check{
		Name:       "test",
		Importance: Required,
		Perform: func() CheckResult {
			return CheckResult{
				Status: status,
			}
		},
	}

	check.WithCache(1)            // sets cache refresh to every 1 second
	defer check.StopCacheTicker() // ensure cleanup

	// initial check should pass
	res := check.getStatus()
	assert.Equal(t, StatusPass, res.Status)

	// Change status after cache should have updated
	status = StatusFail // update status to fail

	// even tought status has changed, we should get the cached value
	res = check.getStatus()
	assert.Equal(t, StatusPass, res.Status)

	// Allow some time for the cache to refresh and then test
	time.Sleep(1100 * time.Millisecond) // wait another tick period
	res = check.getStatus()
	assert.Equal(t, StatusFail, res.Status)
}

func TestStopCacheTicker(t *testing.T) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	stopChan := make(chan struct{})

	check := Check{
		caching: &caching{
			ticker:   ticker,
			stopChan: stopChan,
		},
	}

	done := make(chan bool)
	go func() {
		<-stopChan
		done <- true
	}()

	check.StopCacheTicker()

	select {
	case <-done:
		// Success: channel was closed
	case <-time.After(1 * time.Second):
		t.Error("Stop channel was not closed")
	}
}

func TestUpdateCache(t *testing.T) {
	t.Run("no cached check", func(t *testing.T) {
		check := Check{}
		responseStatus := StatusPass

		check.Perform = func() CheckResult {
			return CheckResult{
				Status: responseStatus,
			}
		}
		status := check.getStatus()
		assert.Equal(t, StatusPass, status.Status)
		check.UpdateCache()

		// update status and check that we do not get a cached one
		responseStatus = StatusFail
		status = check.getStatus()
		assert.Equal(t, StatusFail, status.Status)
	})

	t.Run("caching enabled", func(t *testing.T) {
		responseStatus := StatusPass
		c := Check{
			Name: "test",
			Perform: func() CheckResult {
				return CheckResult{
					Status: responseStatus,
				}
			},
		}
		// update every 10 seconds
		c.WithCache(10)

		res := c.getStatus()
		assert.Equal(t, StatusPass, res.Status)

		responseStatus = StatusFail
		// force a cache update
		c.UpdateCache()

		res = c.getStatus()
		assert.Equal(t, StatusFail, res.Status)
	})

}

func TestGetStatus(t *testing.T) {
	n := 0
	c := Check{
		Name: "test",
		Perform: func() CheckResult {
			n++
			return CheckResult{
				Name:   "test",
				Status: StatusPass,
			}
		},
	}

	t.Run("not caching enabled", func(t *testing.T) {
		c.getStatus()
		c.getStatus()
		assert.Equal(t, 2, n)
	})

	t.Run("caching enabled", func(t *testing.T) {
		n = 0
		c.WithCache(100)

		c.getStatus()
		assert.Equal(t, 1, n)
		c.getStatus()
		// it should be still 1 as we get the cached result
		assert.Equal(t, 1, n)

		t.Run("cache empty", func(t *testing.T) {
			// lets force it and remove the cache
			n = 0
			c.caching.cache.Delete("test")
			c.getStatus()
			assert.Equal(t, 1, n)
		})
	})
}
