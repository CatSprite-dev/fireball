package ui

import (
	"slices"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/load"
	"github.com/CatSprite-dev/fireball/v2/internal/money"
	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/CatSprite-dev/fireball/v2/internal/snapshot"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
)

func TestPortfolioValue(t *testing.T) {
	msk := time.FixedZone("MSK", 3*60*60)
	rub := func(amount string) money.Money {
		return money.Money{Amount: decimal.RequireFromString(amount), Currency: "RUB"}
	}
	event := func(kind ledger.OperationKind, asset, amount, qty string, day int) ledger.Event {
		e := ledger.Event{
			AccountID:     "acc-1",
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

	got, err := portfolioValue(p, m, "RUB", day, []string{"acc-1"})
	if err != nil {
		t.Fatal(err)
	}

	// 10 × 110 + 5 × 190 + рубли 100000 − 2000
	want := decimal.RequireFromString("100050")
	if !got.Equal(want) {
		t.Errorf("value = %s, want %s", got, want)
	}

	t.Run("выбран один счёт из двух", func(t *testing.T) {
		// второй счёт: пополнение 50000 и покупка 20 шт. «c» за 4000.
		second := []ledger.Event{
			event(ledger.KindDeposit, "", "50000", "", 2),
			event(ledger.KindBuy, "c", "-4000", "20", 3),
		}
		for i := range second {
			second[i].AccountID = "acc-2"
		}
		both := &load.Portfolio{Events: append(slices.Clone(p.Events), second...)}
		m := &load.Market{Prices: map[string]series.Series{"a": price("110"), "b": price("190"), "c": price("250")}}

		// только первый счёт: те же 100050, позиция «c» и деньги второго счёта не входят.
		got, err := portfolioValue(both, m, "RUB", day, []string{"acc-1"})
		if err != nil {
			t.Fatal(err)
		}
		if !got.Equal(decimal.RequireFromString("100050")) {
			t.Errorf("acc-1 value = %s, want 100050", got)
		}

		// только второй счёт: 20 × 250 + рубли 50000 − 4000 = 51000.
		got, err = portfolioValue(both, m, "RUB", day, []string{"acc-2"})
		if err != nil {
			t.Fatal(err)
		}
		if !got.Equal(decimal.RequireFromString("51000")) {
			t.Errorf("acc-2 value = %s, want 51000", got)
		}
	})

	t.Run("пустой выбор — 0", func(t *testing.T) {
		got, err := portfolioValue(p, m, "RUB", day, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !got.IsZero() {
			t.Errorf("value = %s, want 0", got)
		}
	})

	t.Run("нет цены — ошибка", func(t *testing.T) {
		m := &load.Market{Prices: map[string]series.Series{"a": price("110")}}
		if _, err := portfolioValue(p, m, "RUB", day, []string{"acc-1"}); err == nil {
			t.Error("want error")
		}
	})
}

func TestToAccounts(t *testing.T) {
	accounts := []load.Account{
		{Account: tinvest.Account{ID: "1", Name: "Брокерский", Status: "ACCOUNT_STATUS_OPEN"}},
		{Account: tinvest.Account{ID: "2", Name: "ИИС", Status: "ACCOUNT_STATUS_CLOSED"}},
	}

	got := toAccounts(accounts)

	want := []Account{
		{ID: "1", Name: "Брокерский", Closed: false},
		{ID: "2", Name: "ИИС", Closed: true},
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
