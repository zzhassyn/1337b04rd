package domain

import "time"

type Comment struct {
	ID, PostID                                        int64
	ReplyToID                                         *int64
	Content, ImageURL, UserName, AvatarURL, SessionID string
	CreatedAt                                         time.Time
}
