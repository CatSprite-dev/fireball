package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/CatSprite-dev/fireball/internal/domain"
)

type CacheManager struct {
	store        *Store
	portfolioTTL time.Duration
	chartTTL     time.Duration
}

func NewCacheManager(store *Store, portfolioTTL time.Duration, chartTTL time.Duration) *CacheManager {
	return &CacheManager{
		store:        store,
		portfolioTTL: portfolioTTL,
		chartTTL:     chartTTL,
	}
}

func (m *CacheManager) SetPortfolio(ctx context.Context, sessionID string, portfolio domain.Portfolio) error {
	portfolioJSON, err := json.Marshal(portfolio)
	if err != nil {
		return err
	}
	if err := m.store.Set(ctx, "portfolio:"+sessionID, portfolioJSON, m.portfolioTTL); err != nil {
		return fmt.Errorf("failed to store portfolio: %w", err)
	}
	return nil
}

func (m *CacheManager) GetPortfolio(ctx context.Context, sessionID string) (domain.Portfolio, error) {
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

func (m *CacheManager) DeletePortfolio(ctx context.Context, sessionID string) error {
	return m.store.Delete(ctx, "portfolio:"+sessionID)
}

func (m *CacheManager) SetChart(ctx context.Context, sessionID string, period string, index string, chartData domain.ChartData) error {
	portfolioJSON, err := json.Marshal(chartData)
	if err != nil {
		return err
	}
	if err := m.store.Set(ctx, "chart:"+sessionID+":"+period+":"+index, portfolioJSON, m.chartTTL); err != nil {
		return fmt.Errorf("failed to store chart data: %w", err)
	}
	return nil
}

func (m *CacheManager) GetChart(ctx context.Context, sessionID string, period string, index string) (domain.ChartData, error) {
	chartDataJSON, err := m.store.Get(ctx, "chart:"+sessionID+":"+period+":"+index)
	if err != nil {
		return domain.ChartData{}, err
	}

	chartData := domain.ChartData{}
	err = json.Unmarshal([]byte(chartDataJSON), &chartData)
	if err != nil {
		return domain.ChartData{}, err
	}

	return chartData, nil
}

func (m *CacheManager) DeleteChartCache(ctx context.Context, sessionID string) error {
	return m.store.DeleteByPattern(ctx, "chart:"+sessionID+":*")
}
