package http

import (
	"1337b04rd/internal/domain"
	"log/slog"
	"net/http"
	"time"
)

func NewRouter(svc domain.PostService, renderer *Renderer, log *slog.Logger) http.Handler {
	h := NewHandler(svc, renderer, log)
	mux := http.NewServeMux()

	sessionMW := SessionMiddleware(svc, log)
	logMW := loggingMiddleware(log)

	mux.Handle("Get /", sessionMW(http.HandlerFunc(h.CatalogPage)))
	mux.Handle("Get /archive", sessionMW(http.HandlerFunc(h.ArchivePage)))
	mux.Handle("Get /archive/", sessionMW(http.HandlerFunc(h.ArchivePostPage)))
	mux.Handle("Get /create-post", sessionMW(http.HandlerFunc(h.CreatePostPage)))
	mux.Handle("POST /submit-post", sessionMW(http.HandlerFunc(h.SubmitPost)))

	return logMW(mux)
}

func loggingMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			log.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.status,
				"duration", time.Since(start).String(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}
