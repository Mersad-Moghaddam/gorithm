package server

import (
	"bytes"
	"compress/gzip"
	"io/fs"
	"net/http"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/web"
)

type Config struct {
	Addr string
}

type staticFile struct {
	data        []byte
	gzipped     []byte
	contentType string
}

func NewHandler() (http.Handler, error) {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	subFS, err := fs.Sub(web.StaticFS, "static")
	if err != nil {
		return nil, err
	}

	// Pre-load and pre-compress static assets into memory
	staticFiles := make(map[string]staticFile)
	err = fs.WalkDir(subFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(subFS, p)
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		gw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
		if err == nil {
			_, _ = gw.Write(data)
			_ = gw.Close()
		}

		ctype := "application/octet-stream"
		switch {
		case strings.HasSuffix(p, ".html"):
			ctype = "text/html; charset=utf-8"
		case strings.HasSuffix(p, ".css"):
			ctype = "text/css; charset=utf-8"
		case strings.HasSuffix(p, ".js"):
			ctype = "application/javascript; charset=utf-8"
		case strings.HasSuffix(p, ".wasm"):
			ctype = "application/wasm"
		case strings.HasSuffix(p, ".svg"):
			ctype = "image/svg+xml"
		case strings.HasSuffix(p, ".ico"):
			ctype = "image/x-icon"
		}

		staticFiles[p] = staticFile{
			data:        data,
			gzipped:     buf.Bytes(),
			contentType: ctype,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		reqPath := r.URL.Path
		var fileKey string
		if reqPath == "/" || !strings.Contains(reqPath, ".") {
			fileKey = "index.html"
		} else {
			fileKey = strings.TrimPrefix(reqPath, "/static/")
			fileKey = strings.TrimPrefix(fileKey, "/")
		}

		sf, ok := staticFiles[fileKey]
		if !ok {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", sf.contentType)
		if fileKey == "index.html" {
			w.Header().Set("Cache-Control", "no-cache")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}

		acceptsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
		if acceptsGzip && len(sf.gzipped) > 0 && len(sf.gzipped) < len(sf.data) {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sf.gzipped)
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sf.data)
		}
	})

	return mux, nil
}
