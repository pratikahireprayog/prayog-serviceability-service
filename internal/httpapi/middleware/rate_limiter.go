package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter is a simple rate limiting middleware.
func RateLimiter(next http.Handler) http.Handler {
	var mu sync.Mutex
	lastSeen := make(map[string]time.Time)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get IP address for rate limiting
		ip := r.RemoteAddr

		mu.Lock()
		lastTime, exists := lastSeen[ip]
		now := time.Now()

		// Allow 10 requests per second per IP
		if exists && now.Sub(lastTime) < time.Second/10 {
			mu.Unlock()
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Rate limit exceeded"))
			return
		}

		lastSeen[ip] = now
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
