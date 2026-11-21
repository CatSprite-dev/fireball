// Package ui — сервисы и типы для окна. Ядро про него не знает.
package ui

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/load"
	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/CatSprite-dev/fireball/v2/internal/snapshot"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
)

// Overview — сводка портфеля для окна. Деньги — строкой, чтобы не терять копейки во float.
type Overview struct {
	Value string `json:"value"` // стоимость портфеля в базовой валюте, decimal строкой "1044651.48"
	AsOf  string `json:"asOf"`  // день оценки, "2006-01-02"
}

// Account — счёт для окна.
type Account struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Closed bool   `json:"closed"`
}

type PortfolioService struct {
	loader *load.Loader
	base   string
	logger *slog.Logger

	// Временно, до базы данных: данные брокера грузятся один раз при первом вызове.
	// mu не даёт двум вызовам при открытии окна грузить одновременно.
	mu        sync.Mutex
	portfolio *load.Portfolio
	market    *load.Market
	today     time.Time
}

// NewPortfolioService — сервис портфеля для окна поверх загрузчика load.
func NewPortfolioService(loader *load.Loader, base string, logger *slog.Logger) *PortfolioService {
	return &PortfolioService{loader: loader, base: base, logger: logger}
}

// Overview — стоимость выбранных счетов на сегодня. Пустой список — стоимость 0.
func (s *PortfolioService) Overview(ctx context.Context, accountIDs []string) (Overview, error) {
	p, m, today, err := s.data(ctx)
	if err != nil {
		return Overview{}, err
	}
	value, err := portfolioValue(p, m, s.base, today, accountIDs)
	if err != nil {
		return Overview{}, err
	}
	return Overview{Value: value.StringFixed(2), AsOf: today.Format(time.DateOnly)}, nil
}

// Accounts — счета портфеля в порядке, как у брокера.
func (s *PortfolioService) Accounts(ctx context.Context) ([]Account, error) {
	p, _, _, err := s.data(ctx)
	if err != nil {
		return nil, err
	}
	return toAccounts(p.Accounts), nil
}

// data — загруженные портфель и рынок; при первом вызове грузит их. Ошибка не запоминается:
// следующий вызов пробует загрузить снова.
func (s *PortfolioService) data(ctx context.Context) (*load.Portfolio, *load.Market, time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.portfolio != nil {
		return s.portfolio, s.market, s.today, nil
	}

	today := series.Day(time.Now())
	p, err := s.loader.Portfolio(ctx)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	m, err := s.loader.Market(ctx, p, today.AddDate(0, 0, -30), time.Now())
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	s.portfolio, s.market, s.today = p, m, today
	return p, m, today, nil
}

func toAccounts(accounts []load.Account) []Account {
	result := make([]Account, 0, len(accounts))
	for _, a := range accounts {
		result = append(result, Account{
			ID:     a.Account.ID,
			Name:   a.Account.Name,
			Closed: a.Account.Status == string(tinvest.AccountStatusClosed),
		})
	}
	return result
}

// portfolioValue — стоимость позиций и денег счетов accountIDs на день day. Если у позиции нет цены — ошибка:
// без неё сумма молча вышла бы меньше настоящей.
func portfolioValue(p *load.Portfolio, m *load.Market, base string, day time.Time, accountIDs []string) (decimal.Decimal, error) {
	events := ledger.Filter(p.Events, accountIDs)
	positions := ledger.ByInstrument(events, p.Currencies.ByAsset)
	cash := ledger.CashHistory(events, base)
	rows, gaps := snapshot.Build(positions, m.Prices, m.Bonds, cash, base, day, day)
	if len(gaps) > 0 {
		return decimal.Zero, fmt.Errorf("value on %s: no price for %d positions, first %s", day.Format(time.DateOnly), len(gaps), gaps[0].Key)
	}
	total := snapshot.Total(rows, day, day)
	return total[0].Value, nil
}
