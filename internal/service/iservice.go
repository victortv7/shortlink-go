package service

import (
	"context"
	"shortlink-go/internal/model"
)

type IService interface {
	CreateShortLink(ctx context.Context, longURL string) (string, error)
	GetLongURL(ctx context.Context, shortLink string) (string, error)
	GetLinkStats(ctx context.Context, shortLink string) (*model.URL, error)
}
