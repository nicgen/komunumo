package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
)

const (
	CSRFCookieName = "__Host-csrf"
	CSRFHeaderName = "X-CSRF-Token"
	csrfTokenBytes = 32
)

// CSRF implements the double-submit cookie pattern.
//
// Safe methods (GET, HEAD, OPTIONS) pass through and ensure a cookie exists.
// Unsafe methods (POST, PUT, PATCH, DELETE) require the cookie value and the
// X-CSRF-Token header to match in constant time.
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			ensureCSRFCookie(w, r)
			next.ServeHTTP(w, r)
			return
		}

		c, err := r.Cookie(CSRFCookieName)
		if err != nil || c.Value == "" {
			http.Error(w, "CSRF cookie missing", http.StatusForbidden)
			return
		}
		header := r.Header.Get(CSRFHeaderName)
		if header == "" {
			http.Error(w, "CSRF header missing", http.StatusForbidden)
			return
		}
		if subtle.ConstantTimeCompare([]byte(c.Value), []byte(header)) != 1 {
			http.Error(w, "CSRF token mismatch", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ensureCSRFCookie(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(CSRFCookieName); err == nil && c.Value != "" {
		return
	}
	buf := make([]byte, csrfTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		http.Error(w, "csrf: rand failure", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     CSRFCookieName,
		Value:    base64.RawURLEncoding.EncodeToString(buf),
		Path:     "/",
		Secure:   true,
		HttpOnly: false, // readable by JS so the SPA can echo it in the header
		SameSite: http.SameSiteStrictMode,
	})
}
