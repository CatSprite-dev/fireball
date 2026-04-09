package session

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	redisClient *redis.Client
	sessionTTL  time.Duration
}

func NewRedisStore(redisURL string, sessionTTL time.Duration) (*Store, error) {
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
		sessionTTL:  sessionTTL,
	}, nil
}

func (s *Store) Set(ctx context.Context, key string, value any) error {
	return s.redisClient.Set(ctx, key, value, s.sessionTTL).Err()
}

func (s *Store) Get(ctx context.Context, key string) (any, error) {
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
