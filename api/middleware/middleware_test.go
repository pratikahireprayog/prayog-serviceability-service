package middleware

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChain(t *testing.T) {
	// Create handler and middlewares for testing
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware that adds a header
	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Middleware-1", "true")
			next.ServeHTTP(w, r)
		})
	}

	// Create another middleware that adds a different header
	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Middleware-2", "true")
			next.ServeHTTP(w, r)
		})
	}

	// Chain the middlewares
	chainedHandler := Chain(handler, middleware1, middleware2)

	// Create a test request
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	chainedHandler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the headers
	if rr.Header().Get("X-Middleware-1") != "true" {
		t.Errorf("middleware1 was not applied")
	}
	if rr.Header().Get("X-Middleware-2") != "true" {
		t.Errorf("middleware2 was not applied")
	}
}

func TestLogger(t *testing.T) {
	// Create a logger that writes to a string buffer
	var logOutput strings.Builder
	logger := log.New(&logOutput, "", 0)

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Apply logger middleware
	loggerMiddleware := Logger(logger)
	wrappedHandler := loggerMiddleware(handler)

	// Create a test request
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set remote address since httptest doesn't set it
	req.RemoteAddr = "127.0.0.1:1234"

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	wrappedHandler.ServeHTTP(rr, req)

	// Check if something was logged
	if logOutput.Len() == 0 {
		t.Error("Logger middleware did not log anything")
	}

	// Check if the log contains the request method and path
	logStr := logOutput.String()
	if !strings.Contains(logStr, "GET") {
		t.Errorf("Log does not contain request method: %s", logStr)
	}
	if !strings.Contains(logStr, "/test") {
		t.Errorf("Log does not contain request path: %s", logStr)
	}
	if !strings.Contains(logStr, "127.0.0.1") {
		t.Errorf("Log does not contain remote address: %s", logStr)
	}
}

func TestCORS(t *testing.T) {
	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Setup allowed origins
	allowedOrigins := []string{"http://example.com"}
	corsMiddleware := CORS(allowedOrigins)
	wrappedHandler := corsMiddleware(handler)

	// Test cases
	testCases := []struct {
		name                    string
		method                  string
		origin                  string
		expectAllowOriginHeader bool
		expectedStatus          int
	}{
		{
			name:                    "Allowed origin",
			method:                  "GET",
			origin:                  "http://example.com",
			expectAllowOriginHeader: true,
			expectedStatus:          http.StatusOK,
		},
		{
			name:                    "Disallowed origin",
			method:                  "GET",
			origin:                  "http://notallowed.com",
			expectAllowOriginHeader: false,
			expectedStatus:          http.StatusOK,
		},
		{
			name:                    "OPTIONS request",
			method:                  "OPTIONS",
			origin:                  "http://example.com",
			expectAllowOriginHeader: true,
			expectedStatus:          http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, "/test", nil)
			if err != nil {
				t.Fatal(err)
			}

			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}

			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, rr.Code)
			}

			// Check CORS headers
			allowOrigin := rr.Header().Get("Access-Control-Allow-Origin")
			if tc.expectAllowOriginHeader && allowOrigin != tc.origin {
				t.Errorf("Expected Access-Control-Allow-Origin header to be %s, got %s", tc.origin, allowOrigin)
			} else if !tc.expectAllowOriginHeader && allowOrigin != "" {
				t.Errorf("Expected no Access-Control-Allow-Origin header, got %s", allowOrigin)
			}
		})
	}
}

func TestRecovery(t *testing.T) {
	// Create a logger that writes to a string buffer
	var logOutput strings.Builder
	logger := log.New(&logOutput, "", 0)

	// Create a handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Apply recovery middleware
	recoveryMiddleware := Recovery(logger)
	wrappedHandler := recoveryMiddleware(panicHandler)

	// Create a test request
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// This should not panic
	wrappedHandler.ServeHTTP(rr, req)

	// Check status code
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	// Check if panic was logged
	logStr := logOutput.String()
	if !strings.Contains(logStr, "Panic recovered") {
		t.Errorf("Log does not contain panic information: %s", logStr)
	}
}
