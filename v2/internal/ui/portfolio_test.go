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

// allTime — начало периода в нулевой день: все события внутри периода, сводка на начало — нули.
var allTime time.Time

func TestSummarizeValue(t *testing.T) {
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
		}
		if qty != "" {
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

	_, sum, err := summarize(p, m, "RUB", allTime, day, []string{"acc-1"})
	if err != nil {
		t.Fatal(err)
	}

	// 10 × 110 + 5 × 190 + рубли 100000 − 2000
	want := decimal.RequireFromString("100050")
	if !sum.Value.Equal(want) {
		t.Errorf("value = %s, want %s", sum.Value, want)
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
		_, sum, err := summarize(both, m, "RUB", allTime, day, []string{"acc-1"})
		if err != nil {
			t.Fatal(err)
		}
		if !sum.Value.Equal(decimal.RequireFromString("100050")) {
			t.Errorf("acc-1 value = %s, want 100050", sum.Value)
		}

		// только второй счёт: 20 × 250 + рубли 50000 − 4000 = 51000.
		_, sum, err = summarize(both, m, "RUB", allTime, day, []string{"acc-2"})
		if err != nil {
			t.Fatal(err)
		}
		if !sum.Value.Equal(decimal.RequireFromString("51000")) {
			t.Errorf("acc-2 value = %s, want 51000", sum.Value)
		}
	})

	t.Run("пустой выбор — 0", func(t *testing.T) {
		_, sum, err := summarize(p, m, "RUB", allTime, day, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !sum.Value.IsZero() {
			t.Errorf("value = %s, want 0", sum.Value)
		}
	})

	t.Run("нет цены — ошибка", func(t *testing.T) {
		m := &load.Market{Prices: map[string]series.Series{"a": price("110")}}
		if _, _, err := summarize(p, m, "RUB", allTime, day, []string{"acc-1"}); err == nil {
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

func TestSummarizeProfitParts(t *testing.T) {
	msk := time.FixedZone("MSK", 3*60*60)
	at := func(kind ledger.OperationKind, asset, amount, qty string, day int) ledger.Event {
		e := ledger.Event{
			AccountID:     "acc",
			Date:          time.Date(2026, time.March, day, 12, 0, 0, 0, msk),
			OperationKind: kind,
			Amount:        money.Money{Amount: decimal.RequireFromString(amount), Currency: "RUB"},
			BaseAmount:    money.Money{Amount: decimal.RequireFromString(amount), Currency: "RUB"},
		}
		if asset != "" {
			e.AssetUID, e.InstrumentUID, e.InstrumentKind = asset, asset+"-uid", ledger.InstrumentShare
		}
		if qty != "" {
			e.Quantity = decimal.RequireFromString(qty)
		}
		return e
	}
	day := time.Date(2026, time.March, 10, 0, 0, 0, 0, time.UTC)

	p := &load.Portfolio{Events: []ledger.Event{
		at(ledger.KindDeposit, "", "100000", "", 2),
		at(ledger.KindBuy, "a", "-1000", "10", 3), // 10 шт. по 100
		at(ledger.KindFee, "a", "-3", "", 3),      // комиссия за покупку
		at(ledger.KindSell, "a", "480", "4", 4),   // 4 шт. по 120: зафиксировано 480 − 400 = 80
		at(ledger.KindDividend, "a", "60", "", 5),
		at(ledger.KindTax, "a", "-8", "", 5),           // налог с дивиденда
		at(ledger.KindFee, "", "-50", "", 5),           // комиссия за обслуживание, без бумаги
		at(ledger.KindWithdrawal, "", "-20000", "", 6), // вывод
	}}
	m := &load.Market{Prices: map[string]series.Series{
		"a": {{Date: day, Value: decimal.RequireFromString("120")}},
	}}

	_, sum, err := summarize(p, m, "RUB", allTime, day, []string{"acc"})
	if err != nil {
		t.Fatal(err)
	}

	checks := []struct {
		name string
		got  decimal.Decimal
		want string
	}{
		// деньги 100000 − 1000 − 3 + 480 + 60 − 8 − 50 − 20000 = 79479; 6 шт. × 120 = 720.
		{"стоимость", sum.Value, "80199"},
		{"вложено", sum.Invested, "80000"},
		// 80 по продаже и комиссия −3.
		{"зафиксировано", sum.Realized, "77"},
		// 6 шт. стоят 720, куплены по 100 за 600.
		{"бумажная", sum.Paper, "120"},
		{"выплаты", sum.Income, "52"},
		{"на уровне счёта", sum.Account, "-50"},
	}
	for _, c := range checks {
		if !c.got.Equal(decimal.RequireFromString(c.want)) {
			t.Errorf("%s = %s, want %s", c.name, c.got, c.want)
		}
	}

	// 77 + 120 + 52 + 0 − 50 = 199 = 80199 − 80000.
	parts := sum.Realized.Add(sum.Paper).Add(sum.Income).Add(sum.Clearing).Add(sum.Account)
	if profit := sum.Value.Sub(sum.Invested); !parts.Equal(profit) {
		t.Errorf("части прибыли %s, прибыль %s", parts, profit)
	}
}

func TestSummarizeAveragePerAccount(t *testing.T) {
	msk := time.FixedZone("MSK", 3*60*60)
	at := func(account string, kind ledger.OperationKind, asset, amount, qty string, day int) ledger.Event {
		e := ledger.Event{
			AccountID:     account,
			Date:          time.Date(2026, time.March, day, 12, 0, 0, 0, msk),
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
	day := time.Date(2026, time.March, 10, 0, 0, 0, 0, time.UTC)

	// Одна бумага на двух счетах по разной цене: на первом 10 шт. по 100, на втором 10 шт. по 200.
	p := &load.Portfolio{Events: []ledger.Event{
		at("acc-1", ledger.KindDeposit, "", "10000", "", 2),
		at("acc-1", ledger.KindBuy, "a", "-1000", "10", 3),
		at("acc-2", ledger.KindDeposit, "", "10000", "", 2),
		at("acc-2", ledger.KindBuy, "a", "-2000", "10", 3),
		// на втором продали 5 по 150.
		at("acc-2", ledger.KindSell, "a", "750", "5", 4),
	}}
	m := &load.Market{Prices: map[string]series.Series{
		"a": {{Date: day, Value: decimal.RequireFromString("150")}},
	}}

	_, sum, err := summarize(p, m, "RUB", allTime, day, []string{"acc-1", "acc-2"})
	if err != nil {
		t.Fatal(err)
	}

	checks := []struct {
		name string
		got  decimal.Decimal
		want string
	}{
		// по средней второго счёта (200): 750 − 5 × 200 = −250.
		// По общей средней двух счетов (150) вышло бы 0 — так считать нельзя.
		{"зафиксировано", sum.Realized, "-250"},
		// первый: 10 × 150 − 1000 = 500; второй: 5 × 150 − 5 × 200 = −250.
		{"бумажная", sum.Paper, "250"},
		// деньги 9000 + 8750 и бумаги 1500 + 750.
		{"стоимость", sum.Value, "20000"},
		{"вложено", sum.Invested, "20000"},
	}
	for _, c := range checks {
		if !c.got.Equal(decimal.RequireFromString(c.want)) {
			t.Errorf("%s = %s, want %s", c.name, c.got, c.want)
		}
	}
}

func TestSummarizePeriod(t *testing.T) {
	msk := time.FixedZone("MSK", 3*60*60)
	date := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 0, 0, 0, 0, time.UTC) }
	at := func(kind ledger.OperationKind, asset, amount, qty string, day time.Time) ledger.Event {
		e := ledger.Event{
			AccountID:     "acc",
			Date:          time.Date(day.Year(), day.Month(), day.Day(), 12, 0, 0, 0, msk),
			OperationKind: kind,
			Amount:        money.Money{Amount: decimal.RequireFromString(amount), Currency: "RUB"},
			BaseAmount:    money.Money{Amount: decimal.RequireFromString(amount), Currency: "RUB"},
		}
		if asset != "" {
			e.AssetUID, e.InstrumentUID, e.InstrumentKind = asset, asset+"-uid", ledger.InstrumentShare
		}
		if qty != "" {
			e.Quantity = decimal.RequireFromString(qty)
		}
		return e
	}
	from, today := date(time.March, 1), date(time.March, 10)

	p := &load.Portfolio{Events: []ledger.Event{
		// до периода: внесли 10000, купили 10 шт. по 100.
		at(ledger.KindDeposit, "", "10000", "", date(time.January, 10)),
		at(ledger.KindBuy, "a", "-1000", "10", date(time.January, 10)),
		// в периоде: внесли 5000, продали 5 шт. по 130 — зафиксировано 650 − 5 × 100 = 150.
		at(ledger.KindDeposit, "", "5000", "", date(time.March, 5)),
		at(ledger.KindSell, "a", "650", "5", date(time.March, 5)),
		// в периоде: дивиденд 40, налог с него 5.
		at(ledger.KindDividend, "a", "40", "", date(time.March, 6)),
		at(ledger.KindTax, "a", "-5", "", date(time.March, 6)),
	}}
	m := &load.Market{Prices: map[string]series.Series{
		// накануне периода цена 120, сегодня 140.
		"a": {
			{Date: date(time.January, 10), Value: decimal.RequireFromString("100")},
			{Date: date(time.February, 28), Value: decimal.RequireFromString("120")},
			{Date: today, Value: decimal.RequireFromString("140")},
		},
	}}

	end, change, err := summarize(p, m, "RUB", from, today, []string{"acc"})
	if err != nil {
		t.Fatal(err)
	}

	// накануне: деньги 9000 + 10 × 120 = 10200, вложено 10000, бумажная 1200 − 1000 = 200.
	// сегодня: деньги 9000 + 5000 + 650 + 40 − 5 = 14685, бумаги 5 × 140 = 700 — всего 15385;
	// вложено 15000, зафиксировано 150, бумажная 700 − 500 = 200, выплаты 35.
	checks := []struct {
		name string
		got  decimal.Decimal
		want string
	}{
		{"стоимость сегодня", end.Value, "15385"},
		{"вложено за период", change.Invested, "5000"},
		// 15385 − 10200 − 5000.
		{"прибыль за период", change.Value.Sub(change.Invested), "185"},
		{"зафиксировано за период", change.Realized, "150"},
		// бумажная была 200 на 10 шт., стала 200 на 5 шт.: половину прежней бумажной зафиксировали.
		{"бумажная за период", change.Paper, "0"},
		{"выплаты за период", change.Income, "35"},
	}
	for _, c := range checks {
		if !c.got.Equal(decimal.RequireFromString(c.want)) {
			t.Errorf("%s = %s, want %s", c.name, c.got, c.want)
		}
	}
	// 150 + 0 + 35 = 185.
	if parts, profit := change.parts(), change.Value.Sub(change.Invested); !parts.Equal(profit) {
		t.Errorf("части прибыли %s, прибыль %s", parts, profit)
	}
}
