package service

import (
	"context"
	"errors"
	"log"
	"shortlink-go/internal/model"
	"shortlink-go/internal/repository"
	"shortlink-go/pkg/base62"
)

var ErrShortLinkNotFound = errors.New("short link not found")

type Service struct {
	urlRepo repository.URLRepository
}

func NewService(urlRepo repository.URLRepository) *Service {
	return &Service{
		urlRepo: urlRepo,
	}
}

// Inserts a new URL into the database and returns the short link.
func (s *Service) CreateShortLink(ctx context.Context, longURL string) (string, error) {
	id, err := s.urlRepo.CreateShortLink(ctx, longURL)
	if err != nil {
		return "", err
	}
	shortLink := base62.Encode(id)

	// TODO: add Redis

	return shortLink, nil
}

// Retrieves the original URL from the short form and increment the access count.
func (s *Service) GetLongURL(ctx context.Context, shortLink string) (string, error) {
	id := base62.Decode(shortLink)

	var longURL string
	var err error
	// TODO: check Redis first

	// Fetch from database if not found in Redis.
	if longURL == "" {
		longURL, err = s.urlRepo.GetLongURL(ctx, id)
	}

	if err != nil {
		return "", err
	}

	// Increment the access count in the background.
	go func(id int64) {
		bgCtx := context.Background()
		if err := s.urlRepo.IncrementAccessCount(bgCtx, id); err != nil {
			log.Printf("Failed to increment access count for ID %d: %v", id, err)
		}
	}(id)

	return longURL, nil
}

// Returns stats for a given short link.
func (s *Service) GetLinkStats(ctx context.Context, shortLink string) (*model.URL, error) {
	id := base62.Decode(shortLink)
	return s.urlRepo.GetURLStats(ctx, id)
}
