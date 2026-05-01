package domain

import (
	"context"
	"io"
	"time"
)

type PostRepository interface {
	Create(ctx context.Context, post *Post) (*Post, error)
	GetByID(ctx context.Context, id int64) (*Post, error)
	ListActive(ctx context.Context) ([]*Post, error)
	ListAll(ctx context.Context) ([]*Post, error)
	Archive(ctx context.Context, id int64) error
	UpdateLastComment(ctx context.Context, postID int64, t time.Time) error
}

type CommentRepository interface {
	Create(ctx context.Context, comment *Comment) (*Comment, error)
	ListByPostID(ctx context.Context, postID int64) ([]*Comment, error)
}

type SessionRepository interface {
	Create(ctx context.Context, session *UserSession) error
	GetByID(ctx context.Context, id string) (*UserSession, error)
	UpdateName(ctx context.Context, id, name string) error
	Delete(ctx context.Context, id string) error
}

type ImageStore interface {
	UploadPostImage(ctx context.Context, filename string, data io.Reader) (url string, err error)
	UploadCommentImage(ctx context.Context, filename string, data io.Reader) (url string, err error)
}

type AvatarService interface {
	GetAvatar(ctx context.Context, id int) (avatarURL, name string, err error)
	TotalAvatars(ctx context.Context) (int, error)
}

type PostService interface {
	CreatePost(ctx context.Context, title, content, userName string, image io.Reader, filename, sessionID string) (*Post, error)
	GetPost(ctx context.Context, id int64) (*Post, []*Comment, error)
	GetArchivedPost(ctx context.Context, id int64) (*Post, []*Comment, error)
	ListCatalog(ctx context.Context) ([]*Post, error)
	ListArchive(ctx context.Context) ([]*Post, error)
	AddComment(ctx context.Context, postID int64, replyToID *int64, content string, image io.Reader, filename, sessionID string) (*Comment, error)
	GetOrCreateSession(ctx context.Context, sessionID string) (*UserSession, error)
	UpdateUserName(ctx context.Context, sessionID, name string) error
}
