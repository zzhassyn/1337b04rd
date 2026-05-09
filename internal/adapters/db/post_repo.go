package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"1337b04rd/internal/domain"
)

type PostRepo struct {
	db *sql.DB
}

func NewPostRepo(db *sql.DB) *PostRepo {
	return &PostRepo{db: db}
}

func (r *PostRepo) Create(ctx context.Context, post *domain.Post) (*domain.Post, error) {
	query := `
		INSERT INTO posts (title, content, image_url, user_name, avatar_url, session_id, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	row := r.db.QueryRowContext(ctx, query, post.Title, post.Content, post.ImageURL, post.UserName, post.AvatarURL, post.SessionID, domain.StatusActive)

	if err := row.Scan(&post.ID, &post.CreatedAt); err != nil {
		return nil, fmt.Errorf("PostRepo.Create: %w", err)
	}

	post.Status = domain.StatusActive

	return post, nil
}

func (r *PostRepo) GetByID(ctx context.Context, id int64) (*domain.Post, error) {
	query := `
		SELECT id, title, content, image_url, user_name, avatar_url, session_id, status, created_at, last_comment_at
        FROM posts
        WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)

	post, err := scanPost(row)
	if err != nil {
		return nil, fmt.Errorf("PostRepo.GetByID: %w", err)
	}
	return post, nil
}

func (r *PostRepo) ListActive(ctx context.Context) ([]*domain.Post, error) {
	query := `
		SELECT id, title, content, image_url, user_name, avatar_url, session_id, status, created_at, last_comment_at
        FROM posts
        WHERE status = $1 ORDER BY created_at DESC`

	return r.queryPosts(ctx, query, domain.StatusActive)
}

func (r *PostRepo) ListAll(ctx context.Context) ([]*domain.Post, error) {
	query := `
		SELECT id, title, content, image_url, user_name, avatar_url, session_id, status, created_at, last_comment_at
        FROM posts ORDER BY created_at DESC`

	return r.queryPosts(ctx, query)
}

func (r *PostRepo) Archive(ctx context.Context, id int64) error {
	query := `
		UPDATE posts SET status = 'archived' WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("PostRepo.Archive: %w", err)
	}

	return nil
}

func (r *PostRepo) UpdateLastComment(ctx context.Context, postID int64, t time.Time) error {
	query := `
		UPDATE posts SET last_comment_at = $1 WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, t, postID)
	if err != nil {
		return fmt.Errorf("PostRepo.UpdateLastComment: %w", err)
	}

	return nil
}

func (r *PostRepo) queryPosts(ctx context.Context, query string, args ...any) ([]*domain.Post, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("PostRepo.queryPosts: %w", err)
	}
	defer rows.Close()

	var posts []*domain.Post

	for rows.Next() {
		post, err := scanPost(rows)
		if err != nil {
			return nil, fmt.Errorf("PostRepo.queryPosts scan: %w", err)
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("PostRepo.queryPosts rows: %w", err)
	}

	return posts, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanPost(s scanner) (*domain.Post, error) {
	var (
		post                domain.Post
		imageURL, sessionID sql.NullString
		lastCommentAt       sql.NullTime
	)

	err := s.Scan(&post.ID, &post.Title, &post.Content,
		&imageURL, &post.UserName, &post.AvatarURL, &sessionID, &post.Status, &post.CreatedAt, &lastCommentAt)
	if err != nil {
		return nil, err
	}

	if imageURL.Valid {
		post.ImageURL = imageURL.String
	}
	if sessionID.Valid {
		post.SessionID = sessionID.String
	}
	if lastCommentAt.Valid {
		post.LastCommentAt = &lastCommentAt.Time
	}

	return &post, nil
}
