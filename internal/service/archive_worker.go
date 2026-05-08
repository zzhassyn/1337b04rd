package service

import (
	"context"
	"log/slog"
	"time"

	"1337b04rd/internal/domain"
)

const (
	archiveNoCommentAfter   = 10 * time.Minute
	archiveLastCommentAfter = 15 * time.Minute
	workerInterval          = 1 * time.Minute
)

func RunArchiveWorker(ctx context.Context, postRepo domain.PostRepository, log *slog.Logger) {
	ticker := time.NewTicker(workerInterval)
	defer ticker.Stop()

	log.Info("archive worker started", "interval", workerInterval)

	for {
		select {
		case <-ctx.Done():
			log.Info("archive worker stopped")

			return
		case <-ticker.C:
			archivePosts(ctx, postRepo, log)
		}
	}
}

func archivePosts(ctx context.Context, postRepo domain.PostRepository, log *slog.Logger) {
	posts, err := postRepo.ListActive(ctx)
	if err != nil {
		log.Error("archive worker: failed to list active posts", "err", err)

		return
	}

	now := time.Now()

	for _, p := range posts {
		shouldArchive := false

		if p.LastCommentAt == nil {
			if now.Sub(p.CreatedAt) >= archiveNoCommentAfter {
				shouldArchive = true
			}
		} else {
			if now.Sub(*p.LastCommentAt) >= archiveLastCommentAfter {
				shouldArchive = true
			}
		}

		if shouldArchive {
			if err := postRepo.Archive(ctx, p.ID); err != nil {
				log.Error("archive worker: failed to archive post",
					"post_id", p.ID, "err", err)

				continue
			}

			log.Info("archive worker: post archived",
				"post_id", p.ID,
				"title", p.Title,
				"created_at", p.CreatedAt,
				"last_comment_at", p.LastCommentAt,
			)
		}
	}
}
