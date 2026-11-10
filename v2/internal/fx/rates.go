package fx

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/money"
	"github.com/shopspring/decimal"
)

var (
	ErrUnknownCurrency = errors.New("fx: unknown currency")
	ErrNoRate          = errors.New("fx: no rate for date")
)

const baseCurrency = "RUB"

var moscow = time.FixedZone("MSK", 3*60*60)

func tradingDay(t time.Time) time.Time {
	y, m, d := t.In(moscow).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

type Point struct {
	Date time.Time
	Rate decimal.Decimal
}

type Rates struct {
	points map[string][]Point
}

func New(points map[string][]Point) *Rates {
	sorted := make(map[string][]Point, len(points))
	for currency, list := range points {
		list = slices.Clone(list)
		slices.SortFunc(list, func(a, b Point) int { return a.Date.Compare(b.Date) })
		sorted[currency] = list
	}
	return &Rates{points: sorted}
}

func (r *Rates) Rate(currency string, date time.Time) (decimal.Decimal, error) {
	day := tradingDay(date)
	if currency == baseCurrency {
		return decimal.NewFromInt(1), nil
	}

	list, ok := r.points[currency]
	if !ok {
		return decimal.Zero, fmt.Errorf("%w: %s", ErrUnknownCurrency, currency)
	}

	var found *Point
	for i := range list {
		if list[i].Date.After(day) {
			break
		}
		found = &list[i]
	}

	if found == nil {
		return decimal.Zero, fmt.Errorf("%w: %s on %s", ErrNoRate, currency, date.Format(time.DateOnly))
	}
	return found.Rate, nil
}

func (r *Rates) ToRUB(m money.Money, date time.Time) (money.Money, error) {
	rate, err := r.Rate(m.Currency, date)
	if err != nil {
		return money.Money{}, err
	}
	return money.Money{Amount: m.Amount.Mul(rate), Currency: baseCurrency}, nil
}
