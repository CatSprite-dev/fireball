package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	redisClient *redis.Client
}

var ErrNotFound = errors.New("key not found")

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
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("failed to get key: %w", err)
	}

	return val, nil
}

func (s *Store) Delete(ctx context.Context, key string) error {
	return s.redisClient.Del(ctx, key).Err()
}

func (s *Store) DeleteByPattern(ctx context.Context, pattern string) error {
	scanCmd := s.redisClient.Scan(ctx, 0, pattern, 0)
	err := scanCmd.Err()
	if err != nil {
		return err
	}
	iter := scanCmd.Iterator()
	for iter.Next(ctx) {
		err := s.redisClient.Unlink(ctx, iter.Val()).Err()
		if err != nil {
			return err
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	return nil
}
