package db

import (
	"1337b04rd/internal/domain"
	"context"
	"database/sql"
	"fmt"
)

type CommentRepo struct {
	db *sql.DB
}

func NewCommentRepo(db *sql.DB) *CommentRepo {
	return &CommentRepo{db: db}
}

func (r *CommentRepo) Create(ctx context.Context, comment *domain.Comment) (*domain.Comment, error) {
	query := `
		INSERT INTO comments (post_id, reply_to_id, content, image_url, user_name, avatar_url, session_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	var replyToID sql.NullInt64
	if comment.ReplyToID != nil {
		replyToID = sql.NullInt64{Int64: *comment.ReplyToID, Valid: true}
	}
	row := r.db.QueryRowContext(ctx, query, comment.PostID, replyToID, comment.Content, comment.ImageURL, comment.UserName, comment.AvatarURL, comment.SessionID)
	if err := row.Scan(&comment.ID, &comment.CreatedAt); err != nil {
		return nil, fmt.Errorf("CommentRepo.Create: %w", err)
	}

	return comment, nil
}

func (r *CommentRepo) ListByPostID(ctx context.Context, postID int64) ([]*domain.Comment, error) {
	query := `
		SELECT id, post_id, reply_to_id, content, image_url, user_name, avatar_url, session_id, created_at
		FROM comments
		WHERE post_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, fmt.Errorf("CommentRepo.ListByPostID: %w", err)
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		comment, err := scanComment(rows)
		if err != nil {
			return nil, fmt.Errorf("CommentRepo.ListByPostID scan: %w", err)
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("CommentRepo.ListByPostID rows: %w", err)
	}

	return comments, nil
}

func scanComment(s scanner) (*domain.Comment, error) {
	var (
		c         domain.Comment
		replyToID sql.NullInt64
		imageURL  sql.NullString
		sessionID sql.NullString
	)

	err := s.Scan(
		&c.ID, &c.PostID, &replyToID, &c.Content, &imageURL,
		&c.UserName, &c.AvatarURL, &sessionID, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if replyToID.Valid {
		id := replyToID.Int64
		c.ReplyToID = &id
	}

	if imageURL.Valid {
		c.ImageURL = imageURL.String
	}

	if sessionID.Valid {
		c.SessionID = sessionID.String
	}

	return &c, nil
}
