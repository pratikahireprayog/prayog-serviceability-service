package http

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a test logger
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

	// Test cases
	tests := []struct {
		name        string
		addr        string
		options     []ServerOption
		expectError bool
	}{
		{
			name:        "Default options",
			addr:        ":8080",
			options:     []ServerOption{},
			expectError: false,
		},
		{
			name:        "With logger",
			addr:        ":8080",
			options:     []ServerOption{WithLogger(logger)},
			expectError: false,
		},
		{
			name: "With all timeouts",
			addr: ":8080",
			options: []ServerOption{
				WithReadTimeout(5 * time.Second),
				WithWriteTimeout(5 * time.Second),
				WithIdleTimeout(60 * time.Second),
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create server with test options
			server := NewServer(tc.addr, handler, tc.options...)

			// Check if server was created successfully
			if server == nil {
				t.Fatal("Server should not be nil")
			}

			// Check server address
			if server.server.Addr != tc.addr {
				t.Errorf("Expected server address to be %s, got %s", tc.addr, server.server.Addr)
			}
		})
	}
}

func TestServerShutdown(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Create our custom server with the test server's handler
	server := NewServer(":8081", ts.Config.Handler)

	// Start the server in a goroutine
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server failed to start: %v", err)
		}
	}()

	// Allow the server time to start
	time.Sleep(100 * time.Millisecond)

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test shutdown
	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("Server shutdown failed: %v", err)
	}
}
