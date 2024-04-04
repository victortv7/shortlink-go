package service

import (
	"context"
	"errors"
	"log"
	"shortlink-go/internal/cache"
	"shortlink-go/internal/model"
	"shortlink-go/internal/repository"
	"shortlink-go/pkg/base62"
)

const REDIS_KEY_PREFIX = "shortlink:"

var ErrShortLinkNotFound = errors.New("short link not found")

type Service struct {
	urlRepo     repository.URLRepository
	redisClient cache.RedisClient
}

func NewService(urlRepo repository.URLRepository, redisClient cache.RedisClient) *Service {
	return &Service{
		urlRepo:     urlRepo,
		redisClient: redisClient,
	}
}

// Inserts a new URL into the database and returns the short link.
func (s *Service) CreateShortLink(ctx context.Context, longURL string) (string, error) {
	// Insert the long URL into the database and get the ID.
	id, err := s.urlRepo.CreateShortLink(ctx, longURL)
	if err != nil {
		return "", err
	}

	shortLink := base62.Encode(id) // Encode the ID to base62 to get the short link.

	// Cache the long URL in Redis using the short link as the key.
	err = s.redisClient.Set(ctx, REDIS_KEY_PREFIX+shortLink, longURL, 0).Err()
	if err != nil {
		log.Printf("Error caching short link in Redis: %v", err)
	}

	return shortLink, nil
}

// Retrieves the original URL from the short form and increment the access count.
func (s *Service) GetLongURL(ctx context.Context, shortLink string) (string, error) {
	id := base62.Decode(shortLink) // Get the DB ID from the short link.

	// Check if the long URL is present in Redis.
	longURL, err := s.redisClient.Get(ctx, REDIS_KEY_PREFIX+shortLink).Result()

	// Fetch from database if not found.
	if err != nil {
		longURL, err = s.urlRepo.GetLongURL(ctx, id)
		if err != nil {
			return "", err
		}
		// Cache the result in Redis for future requests.
		err = s.redisClient.Set(ctx, "shortlink:"+shortLink, longURL, 0).Err()
		if err != nil {
			log.Printf("Failed to cache short link in Redis: %v", err)
		}
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
