package http

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
)

type Renderer struct {
	templates *template.Template
	log       *slog.Logger
}

func NewRenderer(templateDir string, log *slog.Logger) (*Renderer, error) {
	pattern := filepath.Join(templateDir, "*.html")
	tmpl, err := template.ParseGlob(pattern)
	if err != nil {
		return nil, fmt.Errorf("renderer: parse templates: %w", err)
	}

	return &Renderer{
		templates: tmpl,
		log:       log,
	}, nil
}

func (r *Renderer) Render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := r.templates.ExecuteTemplate(w, name, data); err != nil {
		r.log.Error("renderer: execute template", "name", name, "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (r *Renderer) RenderError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	r.Render(w, "error.html", map[string]any{
		"Code":    code,
		"Message": message,
	})
}
