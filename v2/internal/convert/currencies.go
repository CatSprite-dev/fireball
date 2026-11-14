package convert

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
	"github.com/shopspring/decimal"
)

var ErrAmbiguousCurrency = errors.New("convert: several exchange instruments for currency")

const exchangeCurrencyClass = "CETS"

type CurrencyInstrument struct {
	UID     string
	Ticker  string
	Nominal decimal.Decimal
}

func ExchangeCurrencies(list []tinvest.Currency) (map[string]CurrencyInstrument, error) {
	result := make(map[string]CurrencyInstrument)

	for _, c := range list {
		if c.ClassCode != exchangeCurrencyClass || !strings.HasSuffix(c.Ticker, "TOM") {
			continue
		}

		currency := strings.ToUpper(c.IsoCurrencyName)
		if currency == "RUB" {
			continue
		}

		if existing, ok := result[currency]; ok {
			return nil, fmt.Errorf("%w: %s (%s, %s)", ErrAmbiguousCurrency, currency, existing.Ticker, c.Ticker)
		}

		nominal, err := ToDecimal(c.Nominal.Units, c.Nominal.Nano)
		if err != nil {
			return nil, fmt.Errorf("currency %s: nominal: %w", currency, err)
		}

		result[currency] = CurrencyInstrument{UID: c.UID, Ticker: c.Ticker, Nominal: nominal}
	}

	return result, nil
}

func CurrencyByAsset(currencies []tinvest.Currency) map[string]string {
	result := make(map[string]string, len(currencies))
	for _, c := range currencies {
		code := strings.ToUpper(c.IsoCurrencyName)
		if code == "RUB" || c.AssetUID == "" {
			continue
		}
		result[c.AssetUID] = code
	}
	return result
}
