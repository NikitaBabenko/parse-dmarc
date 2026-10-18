// Package httpauth protects the entire dashboard origin in this deployment fork.
package httpauth

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"
	"os"
	"strings"
)

// Wrap fails closed: this fork never serves data without owner-supplied credentials.
func Wrap(next http.Handler) (http.Handler, error) {
	user, password := os.Getenv("HTTP_AUTH_USER"), os.Getenv("HTTP_AUTH_PASSWORD")
	if strings.TrimSpace(user) == "" || strings.ContainsAny(user, ":\r\n") || len(password) < 24 || strings.TrimSpace(password) == "" {
		return nil, errors.New("HTTP_AUTH_USER and a unique HTTP_AUTH_PASSWORD of at least 24 characters are required")
	}
	wantUser, wantPassword := sha256.Sum256([]byte(user)), sha256.Sum256([]byte(password))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		// A constant liveness response contains no report, configuration or metrics.
		if r.URL.Path == "/healthz" && (r.Method == http.MethodGet || r.Method == http.MethodHead) {
			w.WriteHeader(http.StatusOK)
			return
		}
		gotUser, gotPassword, ok := r.BasicAuth()
		userHash, passwordHash := sha256.Sum256([]byte(gotUser)), sha256.Sum256([]byte(gotPassword))
		validUser := subtle.ConstantTimeCompare(userHash[:], wantUser[:])
		validPassword := subtle.ConstantTimeCompare(passwordHash[:], wantPassword[:])
		if !ok || validUser&validPassword != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Parse DMARC", charset="UTF-8"`)
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}), nil
}
