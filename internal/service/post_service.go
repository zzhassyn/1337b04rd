package service

import (
	"1337b04rd/internal/domain"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"time"
)

const sessionCookieTTL = 7 * 24 * time.Hour

type PostService struct {
	posts    domain.PostRepository
	comments domain.CommentRepository
	sessions domain.SessionRepository
	images   domain.ImageStore
	avatars  domain.AvatarService
	log      *slog.Logger
}

func New(posts domain.PostRepository,
	comments domain.CommentRepository,
	sessions domain.SessionRepository,
	images domain.ImageStore,
	avatars domain.AvatarService,
	log *slog.Logger) *PostService {
	return &PostService{
		posts:    posts,
		comments: comments,
		sessions: sessions,
		images:   images,
		avatars:  avatars,
		log:      log,
	}
}

func (s *PostService) GetOrCreateSession(ctx context.Context, sessionID string) (*domain.UserSession, error) {
	if sessionID != "" {
		sess, err := s.sessions.GetByID(ctx, sessionID)
		if err != nil {
			return nil, fmt.Errorf("GetOrCreateSession: %w", err)
		}

		if sess != nil && sess.ExpiresAt.After(time.Now()) {
			return sess, nil
		}
	}

	total, err := s.avatars.TotalAvatars(ctx)
	if err != nil {
		s.log.Warn("GetOrCreateSession: cannot fetch total avatars, using 826", "err", err)
	}

	newID, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("GetOrCreateSession generateID: %w", err)
	}

	avatarID := (idChecksum(newID) % total) + 1

	avatarURL, name, err := s.avatars.GetAvatar(ctx, avatarID)
	if err != nil {
		return nil, fmt.Errorf("GetOrCreateSession GetAvatar: %w", err)
	}

	now := time.Now()
	sess := &domain.UserSession{
		ID:        newID,
		AvatarURL: avatarURL,
		AvatarID:  avatarID,
		UserName:  name,
		CreatedAt: now,
		ExpiresAt: now.Add(sessionCookieTTL),
	}

	if err := s.sessions.Create(ctx, sess); err != nil {
		return nil, fmt.Errorf("GetOrCreateSession Create: %w", err)
	}

	s.log.Info("session created", "id", newID, "avatar_id", avatarID, "name", name)

	return sess, nil
}

func (s *PostService) UpdateUserName(ctx context.Context, sessionID, name string) error {
	if err := s.sessions.UpdateName(ctx, sessionID, name); err != nil {
		return fmt.Errorf("UpdateUserName: %w", err)
	}

	return nil
}

func (s *PostService) CreatePost(ctx context.Context, title, content, userName string, image io.Reader, filename, sessionID string) (*domain.Post, error) {
	sess, err := s.sessions.GetByID(ctx, sessionID)
	if err != nil || sess == nil {
		return nil, fmt.Errorf("CreatePost: session not found: %w", err)
	}

	displayName := sess.UserName
	if userName != "" {
		displayName = userName
	}

	var imageURL string

	if image != nil && filename != "" {
		objectName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filename)

		imageURL, err = s.images.UploadPostImage(ctx, objectName, image)
		if err != nil {
			return nil, fmt.Errorf("CreatePost upload: %w", err)
		}
	}

	p := &domain.Post{
		Title:     title,
		Content:   content,
		ImageURL:  imageURL,
		UserName:  displayName,
		AvatarURL: sess.AvatarURL,
		SessionID: sessionID,
	}

	created, err := s.posts.Create(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("CreatePost: %w", err)
	}

	s.log.Info("post created", "id", created.ID, "title", title)

	return created, nil
}

func (s *PostService) GetPost(ctx context.Context, id int64) (*domain.Post, []*domain.Comment, error) {
	p, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("GetPost: %w", err)
	}

	if p == nil {
		return nil, nil, nil
	}

	comments, err := s.comments.ListByPostID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("GetPost comments: %w", err)
	}

	return p, comments, nil
}

func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func idChecksum(id string) int {
	var sum int

	for _, c := range id {
		sum += int(c)
	}

	return sum
}
