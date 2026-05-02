package middleware

import (
	"context"
	"net/http"

	"komunumo/backend/internal/ports"
)

type contextKey string

const SessionIDKey contextKey = "session_id"

func Auth(sessions ports.SessionRepository, clock ports.Clock) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil || cookie.Value == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"non authentifié"}`))
				return
			}

			_, err = sessions.FindByID(r.Context(), cookie.Value, clock.Now())
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"non authentifié"}`))
				return
			}

			ctx := context.WithValue(r.Context(), SessionIDKey, cookie.Value)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuth(sessions ports.SessionRepository, clock ports.Clock) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil || cookie.Value == "" {
				next.ServeHTTP(w, r)
				return
			}

			_, err = sessions.FindByID(r.Context(), cookie.Value, clock.Now())
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), SessionIDKey, cookie.Value)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
