package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type visitor struct {
	lastSeen time.Time
	count    int
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.RWMutex
)

// RateLimit limits requests per IP
func RateLimit(requestsPerMinute int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIP(r)

			mu.Lock()
			v, exists := visitors[ip]

			if !exists {
				visitors[ip] = &visitor{
					lastSeen: time.Now(),
					count:    1,
				}
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			if time.Since(v.lastSeen) > time.Minute {
				v.count = 1
				v.lastSeen = time.Now()
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			if v.count >= requestsPerMinute {
				mu.Unlock()
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			v.count++
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

// getIP extracts IP from request
func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// CleanupVisitors periodically cleans old visitor entries
func CleanupVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > 5*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()
}
