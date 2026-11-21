// Package ui — сервисы и типы для окна. Ядро про него не знает.
package ui

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"

	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/load"
	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/CatSprite-dev/fireball/v2/internal/snapshot"
)

// Overview — сводка портфеля для окна. Деньги — строкой, чтобы не терять копейки во float.
type Overview struct {
	Value string `json:"value"` // стоимость портфеля в базовой валюте, decimal строкой "1044651.48"
	AsOf  string `json:"asOf"`  // день оценки, "2006-01-02"
}

type PortfolioService struct {
	loader *load.Loader
	base   string
	logger *slog.Logger
}

// NewPortfolioService — сервис портфеля для окна поверх загрузчика load.
func NewPortfolioService(loader *load.Loader, base string, logger *slog.Logger) *PortfolioService {
	return &PortfolioService{loader: loader, base: base, logger: logger}
}

// Overview — стоимость портфеля на сегодня.
func (s *PortfolioService) Overview(ctx context.Context) (Overview, error) {
	today := series.Day(time.Now())

	p, err := s.loader.Portfolio(ctx)
	if err != nil {
		return Overview{}, err
	}
	m, err := s.loader.Market(ctx, p, today.AddDate(0, 0, -30), time.Now())
	if err != nil {
		return Overview{}, err
	}
	value, err := portfolioValue(p, m, s.base, today)
	if err != nil {
		return Overview{}, err
	}
	return Overview{Value: value.StringFixed(2), AsOf: today.Format(time.DateOnly)}, nil
}

// portfolioValue — стоимость всех позиций и денег на день day. Если у позиции нет цены — ошибка:
// без неё сумма молча вышла бы меньше настоящей.
func portfolioValue(p *load.Portfolio, m *load.Market, base string, day time.Time) (decimal.Decimal, error) {
	positions := ledger.ByInstrument(p.Events, p.Currencies.ByAsset)
	cash := ledger.CashHistory(p.Events, base)
	rows, gaps := snapshot.Build(positions, m.Prices, m.Bonds, cash, base, day, day)
	if len(gaps) > 0 {
		return decimal.Zero, fmt.Errorf("value on %s: no price for %d positions, first %s", day.Format(time.DateOnly), len(gaps), gaps[0].Key)
	}
	total := snapshot.Total(rows, day, day)
	return total[0].Value, nil
}
