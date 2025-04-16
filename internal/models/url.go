package models

import (
	"database/sql"
	"time"
)

type User struct {
    ID        int       `db:"id"`
    Email     string    `db:"email"`
    Password  string    `db:"password"`
    GoogleID  sql.NullString    `db:"google_id"`
    Name      sql.NullString     `db:"name"`
    AvatarURL sql.NullString     `db:"avatar_url"`
    CreatedAt time.Time `db:"created_at"`
}

type URL struct {
	ID        int       `json:"id" db:"id"`
	UserId    int       `json:"user_id" db:"user_id"`
	ShortUrl  string    `json:"short_url" db:"short_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	LongURL   string    `db:"long_url" json:"long_url"`
	Clicks    int       `db:"clicks" json:"clicks"`
}
