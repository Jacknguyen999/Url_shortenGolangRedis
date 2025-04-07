package models

import "time"

type User struct {
	ID        int       `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"password" db:"password"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
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
