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
	portfolio, err := s.cacheManager.GetPortfolio(ctx, sessionID)
	if errors.Is(err, storage.ErrNotFound) {
		portfolio, err = s.calculator.GetFullPortfolio(request)
		if err != nil {
			return domain.Portfolio{}, err
		}
		err = s.cacheManager.SetPortfolio(ctx, sessionID, portfolio)
		if err != nil {
			log.Printf("cannot put portfolio cache for %s", sessionID)
		}
		return portfolio, nil
	} else if err == nil {
		return portfolio, nil
	}
	return domain.Portfolio{}, err
}

func (s *PortfolioService) GetOrFetchChartData(
	ctx context.Context,
	sessionID string,
	token string,
	portfolio domain.Portfolio,
	period string,
	indexTicker string,
	candleSource pkg.CandleSource,
) (domain.ChartData, error) {
	chartData, err := s.cacheManager.GetChart(ctx, sessionID, period, indexTicker)
	if errors.Is(err, storage.ErrNotFound) {
		from, to, candleInterval := PeriodToParams(period)
		chartData, err = s.calculator.GetChartData(token, portfolio, indexTicker, from, to, candleInterval, candleSource)
		if err != nil {
			return domain.ChartData{}, err
		}
		err = s.cacheManager.SetChart(ctx, sessionID, period, indexTicker, chartData)
		if err != nil {
			log.Printf("cannot put chart data cache for %s", sessionID)
		}
		return chartData, nil
	} else if err == nil {
		log.Print("found chart data")
		return chartData, nil
	}
	return domain.ChartData{}, err
}

func PeriodToParams(period string) (time.Time, time.Time, pkg.CandleInterval) {
	now := time.Now().UTC()
	switch period {
	case "7d":
		return now.AddDate(0, 0, -7), now, pkg.CandleInterval4Hour
	case "1M":
		return now.AddDate(0, -1, 0), now, pkg.CandleIntervalDay
	case "3M":
		return now.AddDate(0, -3, 0), now, pkg.CandleIntervalDay
	case "6M":
		return now.AddDate(0, -6, 0), now, pkg.CandleIntervalDay
	case "YTD":
		return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()), now, pkg.CandleIntervalDay
	case "1Y":
		return now.AddDate(-1, 0, 0), now, pkg.CandleIntervalDay
	case "5Y":
		return now.AddDate(-5, 0, 0), now, pkg.CandleIntervalWeek
	case "ALL":
		return now.AddDate(-10, 0, 0), now, pkg.CandleIntervalMonth
	default:
		return now.AddDate(0, -6, 0), now, pkg.CandleIntervalDay
	}
}
