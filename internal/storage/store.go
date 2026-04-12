package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	redisClient *redis.Client
}

func NewRedisStore(redisURL string) (*Store, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("error during parsing of redis url: %w", err)
	}

	rdbClient := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = rdbClient.Ping(ctx).Err()
	if err != nil {
		return nil, fmt.Errorf("failed connection to redis: %w", err)
	}

	return &Store{
		redisClient: rdbClient,
	}, nil
}

func (s *Store) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return s.redisClient.Set(ctx, key, value, ttl).Err()
}

func (s *Store) Get(ctx context.Context, key string) (string, error) {
	val, err := s.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("session not found")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get session: %w", err)
	}

	return val, nil
}

func (s *Store) Delete(ctx context.Context, key string) error {
	return s.redisClient.Del(ctx, key).Err()
}
