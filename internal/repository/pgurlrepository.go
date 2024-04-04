// internal/repository/pgurlrepository.go

package repository

import (
	"context"
	"database/sql"
	"errors"
	"shortlink-go/internal/model"
)

type PGURLRepository struct {
	DB *sql.DB
}

func NewPGURLRepository(db *sql.DB) *PGURLRepository {
	return &PGURLRepository{
		DB: db,
	}
}

func (r *PGURLRepository) CreateShortLink(ctx context.Context, longURL string) (int64, error) {
	var id int64
	err := r.DB.QueryRowContext(ctx, "INSERT INTO urls (long_url, access_count) VALUES ($1, $2) RETURNING id", longURL, 0).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *PGURLRepository) GetLongURL(ctx context.Context, id int64) (string, error) {
	var longURL string
	err := r.DB.QueryRowContext(ctx, "SELECT long_url FROM urls WHERE id = $1", id).Scan(&longURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("no URL found")
		}
		return "", err
	}
	return longURL, nil
}

func (r *PGURLRepository) GetURLStats(ctx context.Context, id int64) (*model.URL, error) {
	var url model.URL
	err := r.DB.QueryRowContext(ctx, "SELECT id, long_url, access_count FROM urls WHERE id = $1", id).Scan(&url.ID, &url.LongURL, &url.AccessCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no URL found")
		}
		return nil, err
	}
	return &url, nil
}

func (r *PGURLRepository) IncrementAccessCount(ctx context.Context, id int64) error {
	_, err := r.DB.ExecContext(ctx, "UPDATE urls SET access_count = access_count + 1 WHERE id = $1", id)
	return err
}
