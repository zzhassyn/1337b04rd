package http

import (
	"1337b04rd/internal/domain"
	"context"
	"log/slog"
	"net/http"
	"time"
)

type contextKey string

const (
	SessionContextKey  contextKey = "session"
	ExportedSessionKey            = SessionContextKey
)

func SessionMiddleware(svc domain.PostService, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var sessionID string

			cookie, err := r.Cookie("session_id")
			if err == nil {
				sessionID = cookie.Value
			}

			sess, err := svc.GetOrCreateSession(r.Context(), sessionID)
			if err != nil {
				log.Error("session middleware: get or create session", "err", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    sess.ID,
				Expires:  time.Now().Add(7 * 24 * time.Hour),
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
			})

			ctx := context.WithValue(r.Context(), SessionContextKey, sess)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func SessionFromContext(ctx context.Context) *domain.UserSession {
	session, _ := ctx.Value(SessionContextKey).(*domain.UserSession)

	return session
}
