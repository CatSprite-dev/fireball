package ledger

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/CatSprite-dev/fireball/v2/internal/money"
)

// checkCost сверяет остаток, его стоимость и зафиксированную прибыль
// и проверяет тождество Realized − Cost = Sold + Bought.
func checkCost(t *testing.T, f *Flows, quantity, cost, realized string) {
	t.Helper()
	checks := []struct {
		name string
		got  decimal.Decimal
		want string
	}{
		{"quantity", f.Quantity, quantity},
		{"cost", f.Cost, cost},
		{"realized", f.Realized, realized},
	}
	for _, c := range checks {
		if !c.got.Equal(decimal.RequireFromString(c.want)) {
			t.Errorf("%s = %s, want %s", c.name, c.got, c.want)
		}
	}
	if left, right := f.Realized.Sub(f.Cost), f.Sold.Add(f.Bought); !left.Equal(right) {
		t.Errorf("Realized − Cost = %s, Sold + Bought = %s", left, right)
	}
}

func TestCostAveragePrice(t *testing.T) {
	steps := []struct {
		name                     string
		event                    Event
		quantity, cost, realized string
	}{
		// 10 шт. по 100.
		{"покупка", ev("b1", "", KindBuy, "s", "-1000", "10"), "10", "1000", "0"},
		// ещё 10 по 120: остаток 20 за 2200, средняя 110.
		{"докупка", ev("b2", "", KindBuy, "s", "-1200", "10"), "20", "2200", "0"},
		// 5 по 120: списано 5 × 110 = 550, прибыль 600 − 550 = 50, остаток 15 за 1650.
		{"частичная продажа", ev("s1", "", KindSell, "s", "600", "5"), "15", "1650", "50"},
		// 15 по 100: списано 1650, прибыль 1500 − 1650 = −150, итого 50 − 150 = −100.
		{"полная продажа", ev("s2", "", KindSell, "s", "1500", "15"), "0", "0", "-100"},
	}
	var events []Event
	for _, s := range steps {
		events = append(events, s.event)
		t.Run(s.name, func(t *testing.T) {
			checkCost(t, ByInstrument(events, nil)["s"], s.quantity, s.cost, s.realized)
		})
	}
}

func TestCostBondAmortizationAndRedemption(t *testing.T) {
	// 10 облигаций за 9800 + НКД 50: в стоимость идёт цена без НКД, 9800.
	buy := ev("buy", "", KindBuy, "bond", "-9850", "10")
	buy.AccruedInterest = money.Money{Amount: decimal.RequireFromString("50"), Currency: "RUB"}
	events := []Event{
		buy,
		// вернули 3000 номинала: стоимость остатка 9800 − 3000 = 6800, количество то же.
		ev("amort", "", KindAmortization, "bond", "3000", ""),
		// погасили остаток номинала 7000: прибыль 7000 − 6800 = 200.
		ev("redeem", "", KindRedemption, "bond", "7000", ""),
	}

	f := ByInstrument(events[:2], nil)["bond"]
	checkCost(t, f, "10", "6800", "0")

	f = ByInstrument(events, nil)["bond"]
	checkCost(t, f, "0", "0", "200")
}

func TestCostShortThroughZero(t *testing.T) {
	events := []Event{
		// шорт 10 по 100: остаток −10, стоимость −1000 (деньги получены).
		ev("short", "", KindSell, "s", "1000", "10"),
		// купили 15 по 80 за 1200: 10 закрывают шорт за 800, прибыль 1000 − 800 = 200;
		// 5 открывают длинный остаток за 400.
		ev("buy", "", KindBuy, "s", "-1200", "15"),
	}

	checkCost(t, ByInstrument(events[:1], nil)["s"], "-10", "-1000", "0")
	checkCost(t, ByInstrument(events, nil)["s"], "5", "400", "200")
}

func TestCostAfterSplit(t *testing.T) {
	const asset = "split-asset"
	withSplit(t, asset, split{date: time.Date(2026, 4, 17, 0, 0, 0, 0, moscow), before: 1, after: 10})
	trade := func(id string, kind OperationKind, amount, qty string, date time.Time) Event {
		e := ev(id, "", kind, "split-uid", amount, qty)
		e.InstrumentKind = InstrumentShare
		e.AssetUID = asset
		e.Date = date
		return e
	}

	events := []Event{
		// до сплита 1 шт. за 3107.10 — после сплита это 10 шт., стоимость та же.
		trade("b1", KindBuy, "-3107.10", "1", time.Date(2025, 4, 10, 12, 0, 0, 0, moscow)),
		// после сплита 5 шт. за 1598: остаток 15 за 4705.10.
		trade("b2", KindBuy, "-1598", "5", time.Date(2026, 5, 10, 12, 0, 0, 0, moscow)),
		// 3 шт. за 960: списано 4705.10 × 3 / 15 = 941.02, прибыль 960 − 941.02 = 18.98.
		trade("s1", KindSell, "960", "3", time.Date(2026, 5, 20, 12, 0, 0, 0, moscow)),
	}

	checkCost(t, ByInstrument(events, nil)[asset], "12", "3764.08", "18.98")
}

func TestCostFuturesStaysZero(t *testing.T) {
	buy := ev("buy", "", KindBuy, "fut", "0", "2")
	buy.InstrumentKind = InstrumentFutures
	margin := ev("vm", "", KindVariationMargin, "fut", "1198", "")
	margin.InstrumentKind = InstrumentFutures
	sell := ev("sell", "", KindSell, "fut", "0", "2")
	sell.InstrumentKind = InstrumentFutures

	f := ByInstrument([]Event{buy, margin, sell}, nil)["fut"]

	checkCost(t, f, "0", "0", "0")
	if !f.Clearing.Equal(decimal.RequireFromString("1198")) {
		t.Errorf("clearing = %s, want 1198", f.Clearing)
	}
}

func TestCostOptionExpiredWorthless(t *testing.T) {
	// опцион куплен за 100 и сгорел: уплаченная премия уходит в зафиксированный убыток.
	buy := ev("buy", "", KindBuy, "opt", "-100", "1")
	buy.InstrumentKind = InstrumentOption
	expired := ev("exp", "", KindExpiration, "opt", "0", "")
	expired.InstrumentKind = InstrumentOption

	checkCost(t, ByInstrument([]Event{buy, expired}, nil)["opt"], "0", "0", "-100")
}
