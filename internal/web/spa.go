// Package web serves the embedded Agon landing SPA.
package web

import (
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"

	spaembed "latere.ai/x/debate/internal/web/spa"
)

// MountSPA registers static-asset handlers (immutable cache for
// /assets/, plain serving for /fonts/ and any built top-level file)
// on mux. It returns false when no real frontend is embedded (only
// the dist/PLACEHOLDER), so callers can fall back to a stub.
func MountSPA(mux *http.ServeMux) bool {
	dist, err := fs.Sub(spaembed.FS, "dist")
	if err != nil {
		slog.Warn("spa: no dist embedded", "err", err)
		return false
	}
	if _, err := fs.Stat(dist, "index.html"); err != nil {
		slog.Info("spa: dist present but no index.html; frontend not built")
		return false
	}
	files := http.FS(dist)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == "" || p == "/" {
			serveIndex(w, dist)
			return
		}
		if strings.HasPrefix(p, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			http.FileServer(files).ServeHTTP(w, r)
			return
		}
		clean := path.Clean(p)
		if _, ferr := fs.Stat(dist, strings.TrimPrefix(clean, "/")); ferr == nil {
			http.FileServer(files).ServeHTTP(w, r)
			return
		}
		serveIndex(w, dist)
	})

	mux.Handle("GET /assets/", handler)
	mux.Handle("GET /fonts/", handler)
	mux.Handle("GET /static/", handler)

	slog.Info("spa: mounted")
	return true
}

// SPAFallback serves index.html for any unmatched GET route so
// client-side routing works on deep links.
func SPAFallback(mux *http.ServeMux) {
	dist, err := fs.Sub(spaembed.FS, "dist")
	if err != nil {
		mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "frontend not built", http.StatusServiceUnavailable)
		})
		return
	}
	if _, err := fs.Stat(dist, "index.html"); err != nil {
		mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "frontend not built", http.StatusServiceUnavailable)
		})
		return
	}
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		serveIndex(w, dist)
	})
}

func serveIndex(w http.ResponseWriter, dist fs.FS) {
	b, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		http.Error(w, "spa unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(b)
}
