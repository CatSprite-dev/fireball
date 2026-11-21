package ui

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/load"
	"github.com/CatSprite-dev/fireball/v2/internal/money"
	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/CatSprite-dev/fireball/v2/internal/snapshot"
)

func TestPortfolioValue(t *testing.T) {
	msk := time.FixedZone("MSK", 3*60*60)
	rub := func(amount string) money.Money {
		return money.Money{Amount: decimal.RequireFromString(amount), Currency: "RUB"}
	}
	event := func(kind ledger.OperationKind, asset, amount, qty string, day int) ledger.Event {
		e := ledger.Event{
			Date:          time.Date(2026, time.March, day, 12, 0, 0, 0, msk),
			OperationKind: kind,
			Amount:        rub(amount),
			BaseAmount:    rub(amount),
		}
		if asset != "" {
			e.AssetUID, e.InstrumentUID, e.InstrumentKind = asset, asset+"-uid", ledger.InstrumentShare
			e.Quantity = decimal.RequireFromString(qty)
		}
		return e
	}
	day := time.Date(2026, time.March, 10, 0, 0, 0, 0, time.UTC)
	price := func(v string) series.Series {
		return series.Series{{Date: day, Value: decimal.RequireFromString(v)}}
	}

	p := &load.Portfolio{Events: []ledger.Event{
		event(ledger.KindDeposit, "", "100000", "", 2),
		event(ledger.KindBuy, "a", "-1000", "10", 3),
		event(ledger.KindBuy, "b", "-1000", "5", 4),
	}}
	m := &load.Market{
		Prices: map[string]series.Series{"a": price("110"), "b": price("190")},
		Bonds:  map[string]snapshot.Bond{},
	}

	got, err := portfolioValue(p, m, "RUB", day)
	if err != nil {
		t.Fatal(err)
	}

	// 10 × 110 + 5 × 190 + рубли 100000 − 2000
	want := decimal.RequireFromString("100050")
	if !got.Equal(want) {
		t.Errorf("value = %s, want %s", got, want)
	}

	t.Run("нет цены — ошибка", func(t *testing.T) {
		m := &load.Market{Prices: map[string]series.Series{"a": price("110")}}
		if _, err := portfolioValue(p, m, "RUB", day); err == nil {
			t.Error("want error")
		}
	})
}
