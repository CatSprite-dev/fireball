package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/CatSprite-dev/fireball/internal/pkg"
	"github.com/CatSprite-dev/fireball/internal/storage"
)

type PortfolioService struct {
	calculator   *Calculator
	cacheManager *storage.CacheManager
}

func NewPortfolioService(calc *Calculator, cm *storage.CacheManager) *PortfolioService {
	return &PortfolioService{
		calculator:   calc,
		cacheManager: cm,
	}
}

func (s *PortfolioService) GetOrFetchPortfolio(ctx context.Context, sessionID string, request PortfolioRequest) (domain.Portfolio, error) {
	portfolio, err := s.cacheManager.Get(ctx, sessionID)
	if errors.Is(err, storage.ErrNotFound) {
		portfolio, err = s.calculator.GetFullPortfolio(request)
		if err != nil {
			return domain.Portfolio{}, err
		}
		err = s.cacheManager.Set(ctx, sessionID, portfolio)
		if err != nil {
			log.Printf("cannot put portfolio cache for %s", sessionID)
		}
		return portfolio, nil
	} else if err == nil {
		return portfolio, nil
	}
	return domain.Portfolio{}, err
}

func (s *PortfolioService) GetChartData(
	token string,
	portfolio domain.Portfolio,
	indexTicker string,
	from time.Time,
	to time.Time,
	candleInterval pkg.CandleInterval,
	candleSource pkg.CandleSource,
) (domain.ChartData, error) {
	return s.calculator.GetChartData(token, portfolio, indexTicker, from, to, candleInterval, candleSource)
}
