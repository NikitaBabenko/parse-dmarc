package httpauth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOriginProtection(t *testing.T) {
	t.Setenv("HTTP_AUTH_USER", "test-owner")
	t.Setenv("HTTP_AUTH_PASSWORD", strings.Repeat("x", 32))
	h, err := Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/", "/api/reports", "/api/reports/1", "/api/statistics", "/api/top-sources", "/metrics", "/assets/app.js", "/admin", "/healthz/", "/healthz/../api/reports"} {
		for _, method := range []string{"GET", "HEAD", "OPTIONS", "POST"} {
			for _, credential := range []string{"anonymous", "wrong", "valid"} {
				r := httptest.NewRequest(method, path, nil)
				want := 401
				if credential == "wrong" {
					r.SetBasicAuth("test-owner", "wrong")
				}
				if credential == "valid" {
					r.SetBasicAuth("test-owner", strings.Repeat("x", 32))
					want = 204
				}
				w := httptest.NewRecorder()
				h.ServeHTTP(w, r)
				if w.Code != want {
					t.Errorf("%s %s %s: got %d want %d", method, path, credential, w.Code, want)
				}
				if want == 401 && (w.Header().Get("WWW-Authenticate") == "" || strings.Contains(w.Body.String(), "x")) {
					t.Fatal("invalid unauthorized response")
				}
			}
		}
	}
	for _, method := range []string{"GET", "HEAD", "POST"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(method, "/healthz", nil))
		want := 200
		if method == "POST" {
			want = 401
		}
		if w.Code != want {
			t.Fatalf("health %s: %d", method, w.Code)
		}
	}
}

func TestMissingOrInvalidCredentialsFailClosed(t *testing.T) {
	for _, pair := range [][2]string{{"", ""}, {"owner", "short"}, {"", strings.Repeat("x", 32)}, {"a:b", strings.Repeat("x", 32)}, {"owner", strings.Repeat(" ", 32)}} {
		t.Setenv("HTTP_AUTH_USER", pair[0])
		t.Setenv("HTTP_AUTH_PASSWORD", pair[1])
		if _, err := Wrap(http.NotFoundHandler()); err == nil {
			t.Fatal("invalid credentials accepted")
		}
	}
}
