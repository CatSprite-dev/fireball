package money

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

var ErrCurrencyMismatch = errors.New("money: currency mismatch")

type Money struct {
	Amount   decimal.Decimal
	Currency string
}

func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("%w: %s vs %s", ErrCurrencyMismatch, m.Currency, other.Currency)
	}
	return Money{
		Amount:   m.Amount.Add(other.Amount),
		Currency: m.Currency,
	}, nil
}
