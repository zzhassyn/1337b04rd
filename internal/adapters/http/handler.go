package http

import (
	"1337b04rd/internal/domain"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
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

func (h *Handler) ArchivePostPage(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path, "/archive/")
	if err != nil {
		h.log.Error("ArchivePostPage", "error", err, "post_id", id)
		h.renderer.RenderError(w, http.StatusNotFound, "Post not found")
		return
	}

	post, comments, err := h.svc.GetArchivedPost(r.Context(), id)
	if err != nil {
		h.log.Error("ArchivePostPage", "error", err, "post_id", id)
		h.renderer.RenderError(w, http.StatusInternalServerError, "Failed to load archived post")
		return
	}

	if post == nil {
		h.renderer.RenderError(w, http.StatusNotFound, "Archived post not found")
		return
	}

	h.renderer.Render(w, "archive-post.html", map[string]any{
		"Post":     post,
		"Comments": comments,
	})
}

func (h *Handler) CreatePostPage(w http.ResponseWriter, r *http.Request) {
	sess := SessionFromContext(r.Context())
	h.renderer.Render(w, "create-post.html", map[string]any{
		"Session": sess,
	})
}

func (h *Handler) SubmitPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.log.Error("SubmitPost: parse form", "error", err)
		h.renderer.RenderError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	sess := SessionFromContext(r.Context())
	if sess == nil {
		h.renderer.RenderError(w, http.StatusUnauthorized, "No session")
		return
	}

	title := strings.TrimSpace(r.FormValue("subject"))
	content := strings.TrimSpace(r.FormValue("comment"))
	username := strings.TrimSpace(r.FormValue("name"))

	if title == "" || content == "" {
		h.renderer.RenderError(w, http.StatusBadRequest, "All fields are required")
		return
	}

	var (
		imageReader interface{ Read([]byte) (int, error) }
		filename    string
	)

	file, header, err := r.FormFile("file")
	if err == nil {
		defer file.Close()
		imageReader = file
		filename = header.Filename
	}

	post, err := h.svc.CreatePost(r.Context(), title, content, username, imageReader, filename, sess.ID)
	if err != nil {
		h.log.Error("SubmitPost: create post", "error", err)
		h.renderer.RenderError(w, http.StatusInternalServerError, "Failed to create post")
		return
	}

	h.log.Info("Post submitted", "post_id", post.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func extractID(path, prefix string) (int64, error) {
	idStr := strings.TrimPrefix(path, prefix)
	part := strings.SplitN(idStr, "/", 2)[0]
	return strconv.ParseInt(part, 10, 64)
}
