package middleware

import (
	"fmt"
	"net/http"

	"komunumo/backend/internal/ports"
)

// RateLimit returns a middleware that derives a key from (action, client IP)
// and consumes the underlying RateLimiter port. On block, responds 429 with
// Retry-After header.
func RateLimit(action string, l ports.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := fmt.Sprintf("%s:ip:%s", action, clientIP(r))
			ok, retry := l.Allow(r.Context(), key)
			if !ok {
				if retry > 0 {
					seconds := int(retry.Seconds())
					if seconds < 1 {
						seconds = 1
					}
					w.Header().Set("Retry-After", fmt.Sprintf("%d", seconds))
				}
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	// Prefer X-Forwarded-For first hop (Traefik sets it). Fall back to RemoteAddr.
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		for i := 0; i < len(fwd); i++ {
			if fwd[i] == ',' {
				return fwd[:i]
			}
		}
		return fwd
	}
	return r.RemoteAddr
}
