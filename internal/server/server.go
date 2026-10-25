package server

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/web"
)

type Config struct {
	Addr string
}

func NewHandler() (http.Handler, error) {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Static files from embedded FS
	subFS, err := fs.Sub(web.StaticFS, "static")
	if err != nil {
		return nil, err
	}
	fileServer := http.FileServer(http.FS(subFS))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		path := r.URL.Path
		if path == "/" || !strings.Contains(path, ".") {
			// Serve index.html for root and SPA routes
			indexData, err := fs.ReadFile(subFS, "index.html")
			if err != nil {
				http.Error(w, "index.html not found", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(indexData)
			return
		}

		if strings.HasPrefix(path, "/static/") {
			r.URL.Path = strings.TrimPrefix(path, "/static")
		}

		if strings.HasSuffix(r.URL.Path, ".wasm") {
			w.Header().Set("Content-Type", "application/wasm")
		}

		fileServer.ServeHTTP(w, r)
	})

	return mux, nil
}
