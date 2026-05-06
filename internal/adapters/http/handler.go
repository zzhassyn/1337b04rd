package http

import (
	"1337b04rd/internal/domain"
	"log/slog"
	"net/http"
)

type Handler struct {
	svc      domain.PostService
	renderer *Renderer
	log      *slog.Logger
}

func NewHandler(svc domain.PostService, renderer *Renderer, log *slog.Logger) *Handler {
	return &Handler{
		svc:      svc,
		renderer: renderer,
		log:      log,
	}
}

func (h *Handler) CatalogPage(w http.ResponseWriter, r *http.Request) {
	posts, err := h.svc.ListCatalog(r.Context())
	if err != nil {
		h.log.Error("CatalogPage", "error", err)
		h.renderer.RenderError(w, http.StatusInternalServerError, "Failed to load catalog")
		return
	}

	h.renderer.Render(w, "catalog.html", posts)
}

func (h *Handler) ArchivePage(w http.ResponseWriter, r *http.Request) {
	posts, err := h.svc.ListArchive(r.Context())
	if err != nil {
		h.log.Error("ArchivePage", "error", err)
		h.renderer.RenderError(w, http.StatusInternalServerError, "Failed to load archive")
		return
	}

	h.renderer.Render(w, "archive.html", posts)
}
