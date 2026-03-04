package session

/*
const sessionTTL = 24 * time.Hour

type Store struct {
	client *redis.Client
}

func NewStore(redisURL string) (*Store, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}
	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Store{client: client}, nil
}

func (s *Store) Set(ctx context.Context, sessionID string, encryptedToken string) error {
	return s.client.Set(ctx, "session:"+sessionID, encryptedToken, sessionTTL).Err()
}

func (s *Store) Get(ctx context.Context, sessionID string) (string, error) {
	val, err := s.client.Get(ctx, "session:"+sessionID).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("session not found")
	}
	if err != nil {
		return "", fmt.Errorf("failde to get session: %w", err)
	}
	return val, nil
}

func (s *Store) Delete(ctx context.Context, sessionID string) error {
	return s.client.Del(ctx, "session:"+sessionID).Err()
}
*/
