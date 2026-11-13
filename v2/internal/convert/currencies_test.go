package convert

import (
	"errors"
	"testing"

	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
	"github.com/shopspring/decimal"
)

func currency(uid, ticker, class, iso, nominal string) tinvest.Currency {
	return tinvest.Currency{
		UID:             uid,
		Ticker:          ticker,
		ClassCode:       class,
		IsoCurrencyName: iso,
		Nominal:         tinvest.MoneyValue{Currency: iso, Units: nominal},
	}
}

func TestExchangeCurrencies(t *testing.T) {
	list := []tinvest.Currency{
		currency("a22a1263", "USD000UTSTOM", "CETS", "usd", "1"),
		currency("60d776f7", "USDRUB", "EES_CETS", "usd", "1"),
		currency("a72dafc0", "USD000UTSTOM", "CNGDOTC", "usd", "1"),
		currency("c898d684", "USDFIXME", "CETS", "usd", "0"),
		currency("c7c26356", "KZTRUB_TOM", "CETS", "kzt", "100"),
		currency("a92e2e25", "RUB000UTSTOM", "CETS", "rub", "1"),
	}

	got, err := ExchangeCurrencies(list)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("got %d currencies, want 2 (USD, KZT): %+v", len(got), got)
	}
	if got["USD"].UID != "a22a1263" {
		t.Errorf("USD uid = %s, want a22a1263", got["USD"].UID)
	}
	if !got["KZT"].Nominal.Equal(decimal.NewFromInt(100)) {
		t.Errorf("KZT nominal = %s, want 100", got["KZT"].Nominal)
	}
}

func TestExchangeCurrenciesAmbiguous(t *testing.T) {
	list := []tinvest.Currency{
		currency("a", "USD000UTSTOM", "CETS", "usd", "1"),
		currency("b", "USDRUB_TOM", "CETS", "usd", "1"),
	}

	_, err := ExchangeCurrencies(list)
	if !errors.Is(err, ErrAmbiguousCurrency) {
		t.Errorf("got %v, want ErrAmbiguousCurrency", err)
	}
}

func TestCurrencyByAsset(t *testing.T) {
	list := []tinvest.Currency{
		{AssetUID: "57880736", IsoCurrencyName: "usd", ClassCode: "CETS"},
		{AssetUID: "57880736", IsoCurrencyName: "usd", ClassCode: "EES_CETS"},
		{AssetUID: "e1408696", IsoCurrencyName: "eur"},
		{AssetUID: "3fad82df", IsoCurrencyName: "rub"},
		{AssetUID: "", IsoCurrencyName: "kzt"},
	}

	got := CurrencyByAsset(list)

	want := map[string]string{"57880736": "USD", "e1408696": "EUR"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for asset, code := range want {
		if got[asset] != code {
			t.Errorf("asset %s = %q, want %q", asset, got[asset], code)
		}
	}
}
