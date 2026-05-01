package db

import (
	"1337b04rd/internal/domain"
	"context"
	"database/sql"
	"fmt"
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

	var (
		post                domain.Post
		imageURL, sessionID sql.NullString
		lastCommentAt       sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(&post.ID, &post.Title, &post.Content,
		&imageURL, &post.UserName, &post.AvatarURL, &sessionID, &post.Status, &post.CreatedAt, &lastCommentAt)

	if err != nil {
		return nil, fmt.Errorf("PostRepo.GetByID: %w", err)
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
