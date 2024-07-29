package healthcheck

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestNewHealthChecker(t *testing.T) {
	hc := NewHealthChecker()

	if hc == nil {
		t.Fatal("NewHealthChecker returned nil")
	}

	// Check that checks slice is initialized and empty
	if hc.checks == nil {
		t.Error("Expected checks to be initialized, got nil")
	}
	if len(hc.checks) != 0 {
		t.Errorf("Expected checks to be empty, got length %d", len(hc.checks))
	}
}

func TestRegisterCheck(t *testing.T) {
	hc := NewHealthChecker()

	// Define a mock CheckFunc
	mockCheckFunc := func() (HealthStatus, error) {
		return StatusPass, nil
	}

	// Register a new check
	check := hc.RegisterCheck("PingTest", Required, mockCheckFunc)

	// Check that the returned check is correctly set up
	if check.Name != "PingTest" {
		t.Errorf("Expected Name to be 'PingTest', got '%s'", check.Name)
	}
	if check.Importance != Required {
		t.Errorf("Expected Importance to be 'high', got '%s'", check.Importance)
	}

	// Ensure the check is added to the checks slice
	if len(hc.checks) != 1 {
		t.Errorf("Expected checks slice to have 1 element, has %d", len(hc.checks))
	}
	if hc.checks[0] != check {
		t.Errorf("Expected checks[0] to be the registered check, it is not")
	}
}

func TestPerformChecks(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func() *HealthChecker
		expected HealthCheckResponse
	}{
		{
			name: "All Passing",
			setup: func() *HealthChecker {
				hc := &HealthChecker{checks: []*Check{
					{Importance: Required, Perform: func() CheckResult { return CheckResult{Status: StatusPass} }},
					{Importance: Optional, Perform: func() CheckResult { return CheckResult{Status: StatusFail} }},
					{Importance: Info, Perform: func() CheckResult { return CheckResult{Status: StatusFail} }},
				}}
				return hc
			},
			expected: HealthCheckResponse{Status: StatusPass, StatusCode: 200},
		},
		{
			name: "Required Check Fails",
			setup: func() *HealthChecker {
				hc := &HealthChecker{checks: []*Check{
					{Importance: Required, Perform: func() CheckResult { return CheckResult{Status: StatusFail} }},
					{Importance: Optional, Perform: func() CheckResult { return CheckResult{Status: StatusPass} }},
				}}
				return hc
			},
			expected: HealthCheckResponse{Status: StatusFail, StatusCode: http.StatusServiceUnavailable},
		},
		{
			name: "Optional Check Warns",
			setup: func() *HealthChecker {
				hc := &HealthChecker{checks: []*Check{
					{Importance: Optional, Perform: func() CheckResult { return CheckResult{Status: StatusWarn} }},
					{Importance: Required, Perform: func() CheckResult { return CheckResult{Status: StatusPass} }},
				}}
				return hc
			},
			expected: HealthCheckResponse{Status: StatusWarn, StatusCode: http.StatusMultiStatus},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hc := tc.setup()
			response := hc.PerformChecks()
			if response.Status != tc.expected.Status || response.StatusCode != tc.expected.StatusCode {
				t.Errorf("Test %s failed: expected status %s with code %d, got status %s with code %d",
					tc.name, tc.expected.Status, tc.expected.StatusCode, response.Status, response.StatusCode)
			}
		})
	}
}

func TestHTTPHandler(t *testing.T) {
	hc := NewHealthChecker()

	hc.RegisterCheck("test", Required, func() (HealthStatus, error) {
		return StatusPass, nil
	})

	// Create an HTTP request to use with the handler
	req, err := http.NewRequest("GET", "/healthcheck", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := hc.HTTPHandler()
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect
	pattern := `{"status":"pass","status_code":200,"components":\[{"name":"test","status":"pass","observation_ts":"[^"]+"}\]}`
	regex := regexp.MustCompile(pattern)

	if !regex.MatchString(rr.Body.String()) {
		t.Errorf("handler returned unexpected body: got %s", rr.Body.String())
	}

	// Check the content type is set to application/json
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("content type header does not match: got %v want %v", contentType, "application/json")
	}
}
