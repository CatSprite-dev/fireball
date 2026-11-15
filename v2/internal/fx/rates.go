package fx

import (
	"errors"
	"fmt"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/money"
	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/shopspring/decimal"
)

var (
	ErrUnknownCurrency = errors.New("fx: unknown currency")
	ErrNoRate          = errors.New("fx: no rate for date")
)

type Rates struct {
	base   string
	points map[string]series.Series
}

func New(base string, points map[string]series.Series) *Rates {
	sorted := make(map[string]series.Series, len(points))
	for currency, list := range points {
		sorted[currency] = series.Sorted(list)
	}
	return &Rates{base: base, points: sorted}
}

func (r *Rates) Rate(currency string, date time.Time) (decimal.Decimal, error) {
	if currency == r.base {
		return decimal.NewFromInt(1), nil
	}

	list, ok := r.points[currency]
	if !ok {
		return decimal.Zero, fmt.Errorf("%w: %s", ErrUnknownCurrency, currency)
	}

	p, ok := list.At(series.Day(date))
	if !ok {
		return decimal.Zero, fmt.Errorf("%w: %s on %s", ErrNoRate, currency, date.Format(time.DateOnly))
	}
	return p.Value, nil
}

func (r *Rates) ToBase(m money.Money, date time.Time) (money.Money, error) {
	rate, err := r.Rate(m.Currency, date)
	if err != nil {
		return money.Money{}, err
	}
	return money.Money{Amount: m.Amount.Mul(rate), Currency: r.base}, nil
}
