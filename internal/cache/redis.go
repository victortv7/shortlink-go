package cache

import (
	"context"
	"log"
	"shortlink-go/config"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
}

func NewRedisClient(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Password: cfg.RedisPassword, // no password set
		DB:       cfg.RedisDB,       // use default DB
	})

	// Verify connection
	err := rdb.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis successfully.")
	return rdb
}
