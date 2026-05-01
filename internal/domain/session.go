package domain

import "time"

type UserSession struct {
	ID                   string
	AvatarURL            string
	AvatarID             int
	UserName             string
	CreatedAt, ExpiresAt time.Time
}
