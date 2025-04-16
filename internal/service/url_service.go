package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"time"
	"url_shortenn/internal/models"
)

type URLService struct {
	DB    *sqlx.DB
	Redis *redis.Client
}

func NewURLService(db *sqlx.DB, redis *redis.Client) *URLService {
	return &URLService{DB: db, Redis: redis}
}

func (s *URLService) CreateURL(ctx context.Context, LongURL string, userID int, customShort string) (*models.URL, error) {

	var shortURL string

	if customShort != "" {
		exists, _ := s.ShortURLExists(ctx, customShort)
		if exists {
			return nil, fmt.Errorf("custom short url already exists")
		}
		shortURL = customShort
	} else {
		var err error
		shortURL, err = s.generateShortURL()
		if err != nil {
			return nil, err
		}
	}

	url := &models.URL{
		UserId:    userID,
		LongURL:   LongURL,
		ShortUrl:  shortURL,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(0, 1, 0),
	}

	query := `INSERT INTO urls (user_id, long_url, short_url, created_at, expires_at) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := s.DB.QueryRowContext(ctx, query, url.UserId, url.LongURL, url.ShortUrl, url.CreatedAt, url.ExpiresAt).Scan(&url.ID)
	if err != nil {
		return nil, err
	}

	s.Redis.Set(ctx, "url:"+shortURL, LongURL, 3600)

	return url, nil
}

func (s *URLService) ShortURLExists(ctx context.Context, shortURL string) (bool, error) {
	var exists bool
	err := s.DB.QueryRowContext(ctx, "SELECT EXIST(SELECT 1 FROM urls WHERE short_url = $1)", shortURL).Scan(&exists)
	return exists, err
}

func (s *URLService) generateShortURL() (string, error) {
	bypes := make([]byte, 6)

	if _, err := rand.Read(bypes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bypes)[:6], nil
}

func (s *URLService) GetLongURL(ctx context.Context, shortURL string) (string, error) {
	longURL, err := s.Redis.Get(ctx, "url:"+shortURL).Result()
	if err == redis.Nil {
		var url models.URL
		err := s.DB.GetContext(ctx, &url, "SELECT * FROM urls WHERE short_url = $1", shortURL)
		if err != nil {
			return "", err
		}

		s.Redis.Set(ctx, "url:"+shortURL, url.LongURL, time.Hour*24)

		s.DB.ExecContext(ctx, "UPDATE urls SET clicks = clicks + 1 WHERE id = $1", url.ID)

		return url.LongURL, nil
	} else if err != nil {
		return "", err
	}

	go func() {
		var url models.URL
		if err := s.DB.Get(&url, "SELECT id FROM urls WHERE short_url = $1", shortURL); err == nil {
			s.DB.Exec("UPDATE urls SET clicks = clicks + 1 WHERE id = $1", url.ID)
		}
	}()

	return longURL, nil
}
