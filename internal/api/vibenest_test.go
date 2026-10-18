package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dmarcguardhq/parse-dmarc/internal/metrics"
	"github.com/dmarcguardhq/parse-dmarc/internal/storage"
	"github.com/rs/zerolog"
)

func TestVibeNestProtectedOrigin(t *testing.T) {
	t.Setenv("HTTP_AUTH_USER", "test-owner")
	t.Setenv("HTTP_AUTH_PASSWORD", strings.Repeat("x", 32))
	store, err := storage.NewStorage(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	log := zerolog.Nop()
	s := NewServer(store, "127.0.0.1", 0, metrics.New("test", "test", "test"), &log)
	h, err := s.handler()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/", "/api/reports", "/api/reports/1", "/api/statistics", "/api/top-sources", "/metrics", "/admin", "/assets/missing.js"} {
		for _, method := range []string{"GET", "OPTIONS"} {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(method, path, nil))
			if w.Code != 401 {
				t.Fatalf("anonymous %s %s = %d", method, path, w.Code)
			}
		}
	}
	for _, path := range []string{"/", "/api/reports", "/api/statistics", "/api/top-sources", "/metrics"} {
		r := httptest.NewRequest("GET", path, nil)
		r.SetBasicAuth("test-owner", strings.Repeat("x", 32))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatalf("authenticated %s = %d", path, w.Code)
		}
	}
	t.Setenv("HTTP_AUTH_PASSWORD", "")
	if _, err := s.handler(); err == nil {
		t.Fatal("server accepted missing credentials")
	}
}
