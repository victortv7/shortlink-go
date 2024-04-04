package repository

import (
	"context"
	"shortlink-go/internal/model"
)

type URLRepository interface {
	CreateShortLink(ctx context.Context, longURL string) (int64, error)
	GetLongURL(ctx context.Context, id int64) (string, error)
	GetURLStats(ctx context.Context, id int64) (*model.URL, error)
	IncrementAccessCount(ctx context.Context, id int64) error
}
