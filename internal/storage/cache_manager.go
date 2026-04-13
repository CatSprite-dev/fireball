package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/CatSprite-dev/fireball/internal/domain"
)

type CacheManager struct {
	store *Store
	ttl   time.Duration
}

func NewCacheManager(store *Store, ttl time.Duration) *CacheManager {
	return &CacheManager{
		store: store,
		ttl:   ttl,
	}
}

func (m *CacheManager) Set(ctx context.Context, sessionID string, portfolio domain.Portfolio) error {
	portfolioJSON, err := json.Marshal(portfolio)
	if err != nil {
		return err
	}
	if err := m.store.Set(ctx, "portfolio:"+sessionID, portfolioJSON, m.ttl); err != nil {
		return fmt.Errorf("failed to store portfolio: %w", err)
	}
	return nil
}

func (m *CacheManager) Get(ctx context.Context, sessionID string) (domain.Portfolio, error) {
	portfolioJSON, err := m.store.Get(ctx, "portfolio:"+sessionID)
	if err != nil {
		return domain.Portfolio{}, err
	}

	portfolio := domain.Portfolio{}
	err = json.Unmarshal([]byte(portfolioJSON), &portfolio)
	if err != nil {
		return domain.Portfolio{}, err
	}

	return portfolio, nil
}

func (m *CacheManager) Delete(ctx context.Context, sessionID string) error {
	return m.store.Delete(ctx, "portfolio:"+sessionID)
}
