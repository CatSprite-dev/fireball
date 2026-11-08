package convert

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CatSprite-dev/fireball/v2/internal/money"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
	"github.com/shopspring/decimal"
)

var ErrNoValue = errors.New("convert: no value")

func ToDecimal(units string, nano int32) (decimal.Decimal, error) {
	if units == "" {
		return decimal.Zero, ErrNoValue
	}

	amount, err := decimal.NewFromString(units)
	if err != nil {
		return decimal.Zero, fmt.Errorf("convert: invalid units %q: %w", units, err)
	}

	nanoDecimal := decimal.New(int64(nano), -9)
	return amount.Add(nanoDecimal), nil
}

func ToMoney(mv tinvest.MoneyValue) (money.Money, error) {
	amount, err := ToDecimal(mv.Units, mv.Nano)
	if err != nil {
		return money.Money{}, err
	}
	return money.Money{Amount: amount, Currency: strings.ToUpper(mv.Currency)}, nil
}
