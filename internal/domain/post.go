package domain

import "time"

type PostStatus string

const (
	StatusActive   = "active"
	StatusArchived = "archived"
)

type Post struct {
	ID                                                       int64
	Title, Content, ImageURL, UserName, AvatarURL, SessionID string
	Status                                                   PostStatus
	CreatedAt                                                time.Time
	LastCommentAt                                            *time.Time
}
