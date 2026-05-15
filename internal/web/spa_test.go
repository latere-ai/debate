package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// With only dist/PLACEHOLDER embedded (no index.html — the state of
// `go build ./...` and CI without the Bun stage), MountSPA must
// report "not mounted" and SPAFallback must serve a clean 503 rather
// than panic or 500.
func TestSPAWithoutFrontendBuild(t *testing.T) {
	mux := http.NewServeMux()
	if MountSPA(mux) {
		t.Fatal("MountSPA returned true without an embedded index.html")
	}
	SPAFallback(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("GET / = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}
