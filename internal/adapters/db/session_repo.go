package db

import (
	"1337b04rd/internal/domain"
	"context"
	"database/sql"
	"fmt"
)

type SessionRepo struct {
	db *sql.DB
}

func NewSessionRepo(db *sql.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(ctx context.Context, session *domain.UserSession) error {
	query := `
		INSERT INTO sessions (id, avatar_url, avatar_id, user_name, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.AvatarURL, session.AvatarID, session.UserName, session.CreatedAt, session.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("SessionRepo.Create: %w", err)
	}

	return nil
}

func (r *SessionRepo) GetByID(ctx context.Context, id string) (*domain.UserSession, error) {
	query := `
		SELECT id, avatar_url, avatar_id, user_name, created_at, expires_at
		FROM sessions
		WHERE id = $1`

	var session domain.UserSession

	row := r.db.QueryRowContext(ctx, query, id)

	err := row.Scan(&session.ID, &session.AvatarURL, &session.AvatarID, &session.UserName, &session.CreatedAt, &session.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("SessionRepo.GetByID: %w", err)
	}

	return &session, nil
}

func (r *SessionRepo) UpdateName(ctx context.Context, id, name string) error {
	query := `
		UPDATE sessions SET user_name = $1 WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, name, id)
	if err != nil {
		return fmt.Errorf("SessionRepo.UpdateName: %w", err)
	}

	return nil
}

func (r *SessionRepo) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM sessions WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("SessionRepo.Delete: %w", err)
	}

	return nil
}
