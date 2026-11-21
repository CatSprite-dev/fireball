package ui

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/CatSprite-dev/fireball/v2/internal/convert"
	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/load"
	"github.com/CatSprite-dev/fireball/v2/internal/money"
	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
)

func TestCharts(t *testing.T) {
	msk := time.FixedZone("MSK", 3*60*60)
	start := time.Date(2025, time.March, 10, 0, 0, 0, 0, time.UTC)
	closing, half, today := start.AddDate(0, 0, 100), start.AddDate(0, 0, 182), start.AddDate(0, 0, 365)
	at := func(account string, kind ledger.OperationKind, asset, amount, qty string, day time.Time) ledger.Event {
		e := ledger.Event{
			AccountID:     account,
			Date:          time.Date(day.Year(), day.Month(), day.Day(), 12, 0, 0, 0, msk),
			OperationKind: kind,
			Amount:        money.Money{Amount: decimal.RequireFromString(amount), Currency: "RUB"},
			BaseAmount:    money.Money{Amount: decimal.RequireFromString(amount), Currency: "RUB"},
		}
		if asset != "" {
			e.AssetUID, e.InstrumentUID, e.InstrumentKind = asset, asset+"-uid", ledger.InstrumentShare
			e.Quantity = decimal.RequireFromString(qty)
		}
		return e
	}
	pt := func(day time.Time, v string) series.Point {
		return series.Point{Date: day, Value: decimal.RequireFromString(v)}
	}

	p := &load.Portfolio{
		Currencies: convert.Currencies{ByUID: map[string]string{"rub-uid": "RUB"}},
		Accounts: []load.Account{
			// у брокера 605 рублей, у нас 605 — расхождения нет.
			{Account: tinvest.Account{ID: "acc"}, Portfolio: &tinvest.Portfolio{Positions: []tinvest.Position{
				{InstrumentUID: "rub-uid", Quantity: tinvest.Quotation{Units: "605"}},
			}}},
			// закрытый ИИС: у брокера 0, у нас 820 — брокер не отдал налог при закрытии, расхождение −820.
			{Account: tinvest.Account{ID: "iis"}},
		},
		Events: []ledger.Event{
			// в первый день внесли 1000 и купили 10 шт. по 100.
			at("acc", ledger.KindDeposit, "", "1000", "", start),
			at("acc", ledger.KindBuy, "a", "-1000", "10", start),
			// через год продали 5 шт. по 121.
			at("acc", ledger.KindSell, "a", "605", "5", today),
			// ИИС: внесли 1000, при закрытии вывели 180 — последняя операция счёта.
			at("iis", ledger.KindDeposit, "", "1000", "", start),
			at("iis", ledger.KindWithdrawal, "", "-180", "", closing),
		},
	}
	m := &load.Market{Prices: map[string]series.Series{
		// цена 100 → 110 через полгода → 121 через год.
		"a": {pt(start, "100"), pt(half, "110"), pt(today, "121")},
	}}
	// индекс начинается только через полгода: 1000 → 1050.
	index := series.Series{pt(half, "1000"), pt(today, "1050")}
	accounts := []string{"acc", "iis"}

	got, err := charts(p, m, index, "RUB", accounts, "all", today)
	if err != nil {
		t.Fatal(err)
	}

	first := func(points []Point) Point { return points[0] }
	last := func(points []Point) Point { return points[len(points)-1] }
	on := func(points []Point, day time.Time) Point {
		for _, p := range points {
			if p.Date == day.Format(time.DateOnly) {
				return p
			}
		}
		return Point{}
	}
	checks := []struct {
		name      string
		got, want Point
	}{
		// обе линии — с первого дня индекса. TWR от поправленной стоимости: ИИС уже закрыт,
		// 1920 − 820 = 1100 → 605 + 605 + 820 − 820 = 1210, +10%.
		{"twr начало", first(got.TWR), Point{half.Format(time.DateOnly), "0.00"}},
		{"twr конец", last(got.TWR), Point{today.Format(time.DateOnly), "10.00"}},
		{"индекс начало", first(got.Index), Point{half.Format(time.DateOnly), "0.00"}},
		{"индекс конец", last(got.Index), Point{today.Format(time.DateOnly), "5.00"}},
		// стоимость: до закрытия ИИС поправки нет — 1000 + 1000.
		{"стоимость начало", first(got.Value), Point{start.Format(time.DateOnly), "2000.00"}},
		{"стоимость накануне закрытия", on(got.Value, closing.AddDate(0, 0, -1)), Point{closing.AddDate(0, 0, -1).Format(time.DateOnly), "2000.00"}},
		// с дня закрытия −820: 1000 + 820 − 820.
		{"стоимость в день закрытия", on(got.Value, closing), Point{closing.Format(time.DateOnly), "1000.00"}},
		{"стоимость конец", last(got.Value), Point{today.Format(time.DateOnly), "1210.00"}},
		// тень по бумагам с полугода: бумаги 1100 → 1.1 пая индекса.
		{"бумаги начало", first(got.ShadowTrades.Actual), Point{half.Format(time.DateOnly), "1100.00"}},
		{"тень бумаг начало", first(got.ShadowTrades.Shadow), Point{half.Format(time.DateOnly), "1100.00"}},
		// бумаги: 5 × 121. Тень продала паёв на 605: 1.1 × 1050 − 605.
		{"бумаги конец", last(got.ShadowTrades.Actual), Point{today.Format(time.DateOnly), "605.00"}},
		{"тень бумаг конец", last(got.ShadowTrades.Shadow), Point{today.Format(time.DateOnly), "550.00"}},
		// тень по пополнениям с полугода от поправленной стоимости 1100: 1.1 пая индекса.
		{"стоимость тени начало", first(got.ShadowDeposits.Actual), Point{half.Format(time.DateOnly), "1100.00"}},
		{"тень пополнений начало", first(got.ShadowDeposits.Shadow), Point{half.Format(time.DateOnly), "1100.00"}},
		// пополнений после нет: тень 1.1 × 1050 — несуществующие 820 в индекс не вложены.
		{"стоимость тени конец", last(got.ShadowDeposits.Actual), Point{today.Format(time.DateOnly), "1210.00"}},
		{"тень пополнений конец", last(got.ShadowDeposits.Shadow), Point{today.Format(time.DateOnly), "1155.00"}},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %+v, want %+v", c.name, c.got, c.want)
		}
	}

	// разница теней — реальная минус тень на последнюю дату: 605 − 550 и 1210 − 1155.
	if got.ShadowTrades.Diff != "55.00" || got.ShadowDeposits.Diff != "55.00" {
		t.Errorf("разница теней: по бумагам %s, по пополнениям %s, want 55.00", got.ShadowTrades.Diff, got.ShadowDeposits.Diff)
	}

	// поправка не уводит стоимость в минус.
	for _, p := range got.Value {
		if decimal.RequireFromString(p.Value).IsNegative() {
			t.Errorf("стоимость %s: %s < 0", p.Date, p.Value)
		}
	}

	// последние точки совпадают с карточками: стоимость и индекс — сами по себе,
	// рост портфеля с дня сравнения — через «Против индекса».
	perf, err := performance(p, m, index, "RUB", accounts, "all", today)
	if err != nil {
		t.Fatal(err)
	}
	card, err := overview(p, m, "RUB", start, today, accounts, true)
	if err != nil {
		t.Fatal(err)
	}
	// карточка TWR за всё время — тоже от поправленного ряда: в день закрытия ИИС 2000 → 1000 при выводе 180,
	// (1000 + 180) / 2000 = 0.59; дальше +10% и +10%: 0.59 × 1.21 = 0.7139, −28.61%.
	if perf.TWR != "-28.61" {
		t.Errorf("карточка TWR %s, want -28.61", perf.TWR)
	}
	twr, idx := decimal.RequireFromString(last(got.TWR).Value), decimal.RequireFromString(last(got.Index).Value)
	if idx.StringFixed(2) != perf.Index || twr.Sub(idx).StringFixed(2) != perf.VsIndex || last(got.Value).Value != card.Value {
		t.Errorf("charts twr %s, index %s, value %s; cards index %s, vs index %s, value %s",
			twr, idx, last(got.Value).Value, perf.Index, perf.VsIndex, card.Value)
	}
}
