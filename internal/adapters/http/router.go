package http

import (
	"log/slog"
	"net/http"
	"time"

	"1337b04rd/internal/domain"
)

func NewRouter(svc domain.PostService, renderer *Renderer, log *slog.Logger) http.Handler {
	h := NewHandler(svc, renderer, log)
	mux := http.NewServeMux()

	sessionMW := SessionMiddleware(svc, log)
	logMW := loggingMiddleware(log)

	mux.Handle("GET /", sessionMW(http.HandlerFunc(h.CatalogPage)))
	mux.Handle("GET /archive", sessionMW(http.HandlerFunc(h.ArchivePage)))
	mux.Handle("GET /archive/", sessionMW(http.HandlerFunc(h.ArchivePostPage)))
	mux.Handle("GET /create-post", sessionMW(http.HandlerFunc(h.CreatePostPage)))
	mux.Handle("POST /submit-post", sessionMW(http.HandlerFunc(h.SubmitPost)))
	mux.Handle("GET /post/", sessionMW(http.HandlerFunc(routePostGet(h))))
	mux.Handle("POST /post/", sessionMW(http.HandlerFunc(routePostComment(h))))
	mux.Handle("POST /update-name", sessionMW(http.HandlerFunc(h.UpdateName)))

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

func routePostGet(h *Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.PostPage(w, r)
	}
}

func routePostComment(h *Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.SubmitComment(w, r)
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
