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

func TestPositionTable(t *testing.T) {
	msk := time.FixedZone("MSK", 3*60*60)
	date := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 0, 0, 0, 0, time.UTC) }
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
		}
		if qty != "" {
			e.Quantity = decimal.RequireFromString(qty)
		}
		return e
	}
	from, today := date(time.March, 1), date(time.March, 10)

	p := &load.Portfolio{
		Accounts: []load.Account{{
			Account: tinvest.Account{ID: "acc-1"},
			// тикер и имя берутся из последней операции: бумагу переименовали.
			Operations: []tinvest.Operation{
				{InstrumentUID: "a-uid", Ticker: "OLD", Name: "Старое имя", Date: date(time.January, 10)},
				{InstrumentUID: "a-uid", Ticker: "AAA", Name: "Акция А", Date: date(time.March, 5)},
			},
		}},
		Events: []ledger.Event{
			// первый счёт — как в TestSummarizePeriod: прибыль за период 185 (зафикс. 150 + выплаты 35).
			at("acc-1", ledger.KindDeposit, "", "10000", "", date(time.January, 10)),
			at("acc-1", ledger.KindBuy, "a", "-1000", "10", date(time.January, 10)),
			at("acc-1", ledger.KindDeposit, "", "5000", "", date(time.March, 5)),
			at("acc-1", ledger.KindSell, "a", "650", "5", date(time.March, 5)),
			at("acc-1", ledger.KindDividend, "a", "40", "", date(time.March, 6)),
			at("acc-1", ledger.KindTax, "a", "-5", "", date(time.March, 6)),
			// второй счёт: в периоде внесли 2000, купили 10 шт. по 130, комиссия за обслуживание 20.
			at("acc-2", ledger.KindDeposit, "", "2000", "", date(time.March, 3)),
			at("acc-2", ledger.KindBuy, "a", "-1300", "10", date(time.March, 3)),
			at("acc-2", ledger.KindFee, "", "-20", "", date(time.March, 4)),
		},
	}
	m := &load.Market{Prices: map[string]series.Series{
		"a": {
			{Date: date(time.January, 10), Value: decimal.RequireFromString("100")},
			{Date: date(time.February, 28), Value: decimal.RequireFromString("120")},
			{Date: today, Value: decimal.RequireFromString("140")},
		},
	}}

	got, err := positionTable(p, m, "RUB", from, today, []string{"acc-1", "acc-2"}, false)
	if err != nil {
		t.Fatal(err)
	}

	rows := make(map[string]Position)
	for _, r := range got.Positions {
		rows[r.Key] = r
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2 (a and RUB): %+v", len(rows), got.Positions)
	}

	// бумага «a» на двух счетах: 5 + 10 = 15 шт. по 140 = 2100; остаток куплен за 500 + 1300 = 1800.
	// прибыль: первый счёт 185, второй — бумажная 1400 − 1300 = 100; всего 285.
	// стоимость всего: 2100 + рубли (14685 + 680) = 17465; доля 2100 / 17465 = 12.02%.
	want := Position{
		Key: "a", Ticker: "AAA", Name: "Акция А", Kind: "share",
		Quantity: "15", Closed: false, Value: "2100.00", Share: "12.02",
		Invested: "1800.00", Income: "35.00", Profit: "285.00", ProfitPct: "",
	}
	if rows["a"] != want {
		t.Errorf("a:\n got  %+v\n want %+v", rows["a"], want)
	}

	// рубли: первый счёт 9000 + 5000 + 650 + 40 − 5 = 14685, второй 2000 − 1300 − 20 = 680.
	rub := rows["RUB"]
	if rub.Name != "Рубли" || rub.Kind != "currency" || rub.Value != "15365.00" || rub.Share != "87.98" || rub.Profit != "0.00" {
		t.Errorf("RUB: %+v", rub)
	}

	// комиссия без бумаги; 285 − 20 = 265 — это и есть прибыль за период в карточке.
	if got.AccountLevel != "-20.00" {
		t.Errorf("account level = %s, want -20.00", got.AccountLevel)
	}
	_, change, err := summarize(p, m, "RUB", from, today, []string{"acc-1", "acc-2"})
	if err != nil {
		t.Fatal(err)
	}
	if profit := change.Value.Sub(change.Invested); !profit.Equal(decimal.RequireFromString("265")) {
		t.Errorf("overview profit = %s, want 265", profit)
	}
}

func TestPositionTableClosedProfit(t *testing.T) {
	msk := time.FixedZone("MSK", 3*60*60)
	date := func(d int) time.Time { return time.Date(2026, time.March, d, 0, 0, 0, 0, time.UTC) }
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
			e.Quantity = decimal.RequireFromString(qty)
		}
		return e
	}
	from, today := date(1), date(10)
	m := &load.Market{Prices: map[string]series.Series{
		"a": {{Date: date(2), Value: decimal.RequireFromString("100")}},
		"b": {{Date: date(2), Value: decimal.RequireFromString("100")}},
	}}

	// «b» купили по 100 и продали по 130 — закрыта с прибылью 30. «a» купили и держим — не закрыта.
	p := &load.Portfolio{Events: []ledger.Event{
		at(ledger.KindDeposit, "", "1000", "", date(2)),
		at(ledger.KindBuy, "a", "-100", "1", date(2)),
		at(ledger.KindBuy, "b", "-100", "1", date(3)),
		at(ledger.KindSell, "b", "130", "1", date(4)),
	}}
	got, err := positionTable(p, m, "RUB", from, today, []string{"acc"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if got.ClosedProfit != "30.00" {
		t.Errorf("closed profit = %q, want 30.00", got.ClosedProfit)
	}

	// без закрытых позиций — пусто.
	p.Events = p.Events[:2]
	got, err = positionTable(p, m, "RUB", from, today, []string{"acc"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if got.ClosedProfit != "" {
		t.Errorf("closed profit = %q, want empty", got.ClosedProfit)
	}
}

func TestUnexplainedCash(t *testing.T) {
	today := time.Date(2026, time.March, 10, 0, 0, 0, 0, time.UTC)
	deposit := func(account, amount string) ledger.Event {
		return ledger.Event{
			AccountID:     account,
			Date:          time.Date(2026, time.March, 2, 12, 0, 0, 0, time.UTC),
			OperationKind: ledger.KindDeposit,
			Amount:        money.Money{Amount: decimal.RequireFromString(amount), Currency: "RUB"},
			BaseAmount:    money.Money{Amount: decimal.RequireFromString(amount), Currency: "RUB"},
		}
	}
	rubles := func(units string, nano int32) tinvest.Position {
		return tinvest.Position{InstrumentUID: "rub-uid", Quantity: tinvest.Quotation{Units: units, Nano: nano}}
	}
	p := &load.Portfolio{
		Currencies: convert.Currencies{ByUID: map[string]string{"rub-uid": "RUB", "usd-uid": "USD"}},
		Accounts: []load.Account{
			// открытый счёт: у брокера 1000 рублей и доллары, у нас 1000 — расхождения нет, доллары не сравниваются.
			{Account: tinvest.Account{ID: "open"}, Portfolio: &tinvest.Portfolio{Positions: []tinvest.Position{
				rubles("1000", 0),
				{InstrumentUID: "usd-uid", Quantity: tinvest.Quotation{Units: "5"}},
			}}},
			// закрытый счёт: портфеля нет, у брокера 0, а у нас осталось 820 — брокер не отдал операцию.
			{Account: tinvest.Account{ID: "closed"}},
			// открытый: у брокера 499.50, у нас 632.39 — расхождение −132.89.
			{Account: tinvest.Account{ID: "piggy"}, Portfolio: &tinvest.Portfolio{Positions: []tinvest.Position{
				rubles("499", 500_000_000),
			}}},
		},
		Events: []ledger.Event{deposit("open", "1000"), deposit("closed", "820"), deposit("piggy", "632.39")},
	}

	checks := []struct {
		name     string
		accounts []string
		want     string
	}{
		{"все счета", []string{"open", "closed", "piggy"}, "-952.89"},
		{"только открытый без расхождения", []string{"open"}, "0"},
		{"только закрытый", []string{"closed"}, "-820"},
		{"ничего не выбрано", nil, "0"},
	}
	for _, c := range checks {
		got, err := unexplainedCash(p, "RUB", today, c.accounts)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if !got.Equal(decimal.RequireFromString(c.want)) {
			t.Errorf("%s: got %s, want %s", c.name, got, c.want)
		}
	}
}

func TestUnexplainedInCardsAndTable(t *testing.T) {
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
			e.Quantity = decimal.RequireFromString(qty)
		}
		return e
	}
	// portfolio — счёт, у которого брокер показывает brokerRubles рублей; у нас рублей 10000.
	portfolio := func(brokerRubles string) *load.Portfolio {
		return &load.Portfolio{
			Currencies: convert.Currencies{ByUID: map[string]string{"rub-uid": "RUB"}},
			Accounts: []load.Account{{
				Account: tinvest.Account{ID: "acc"},
				Portfolio: &tinvest.Portfolio{Positions: []tinvest.Position{
					{InstrumentUID: "rub-uid", Quantity: tinvest.Quotation{Units: brokerRubles}},
				}},
			}},
			Events: []ledger.Event{
				// 10 января внесли 10000 и купили 10 шт. по 100; 5 марта внесли ещё 1000.
				at(ledger.KindDeposit, "", "10000", "", date(time.January, 10)),
				at(ledger.KindBuy, "a", "-1000", "10", date(time.January, 10)),
				at(ledger.KindDeposit, "", "1000", "", date(time.March, 5)),
			},
		}
	}
	m := &load.Market{Prices: map[string]series.Series{
		"a": {
			{Date: date(time.January, 10), Value: decimal.RequireFromString("100")},
			{Date: date(time.February, 28), Value: decimal.RequireFromString("120")},
			{Date: date(time.March, 10), Value: decimal.RequireFromString("140")},
		},
	}}
	today := date(time.March, 10)
	accounts := []string{"acc"}

	// у брокера на 100 рублей меньше: он не отдал в API операцию на 100 (например, налог).
	withGap := portfolio("9900")

	t.Run("всё время", func(t *testing.T) {
		from := date(time.January, 10)
		card, err := overview(withGap, m, "RUB", from, today, accounts, true)
		if err != nil {
			t.Fatal(err)
		}
		table, err := positionTable(withGap, m, "RUB", from, today, accounts, true)
		if err != nil {
			t.Fatal(err)
		}

		// наша стоимость 10000 + 10 × 140 = 11400, с расхождением 11300 — как у брокера.
		// прибыль 11300 − вложено 11000 = 300 (наша 400 минус расхождение 100).
		if card.Value != "11300.00" || card.Profit != "300.00" {
			t.Errorf("card: value %s, profit %s; want 11300.00, 300.00", card.Value, card.Profit)
		}
		if table.Unexplained != "-100.00" || table.UnexplainedProfit != "-100.00" {
			t.Errorf("unexplained = %q / %q, want -100.00 in both", table.Unexplained, table.UnexplainedProfit)
		}
		if table.TotalProfit != card.Profit || table.TotalValue != card.Value {
			t.Errorf("table totals %s / %s, card %s / %s", table.TotalValue, table.TotalProfit, card.Value, card.Profit)
		}

		// группы: акции — 1400 и прибыль 400, валюта — рубли 10000 без прибыли.
		groups := make(map[string]GroupTotal)
		for _, g := range table.Groups {
			groups[g.Kind] = g
		}
		if groups["share"] != (GroupTotal{Kind: "share", Value: "1400.00", Profit: "400.00"}) ||
			groups["currency"] != (GroupTotal{Kind: "currency", Value: "10000.00", Profit: "0.00"}) {
			t.Errorf("groups = %+v", table.Groups)
		}
	})

	t.Run("период внутри — расхождение не задевает прибыль", func(t *testing.T) {
		from := date(time.March, 1)
		card, err := overview(withGap, m, "RUB", from, today, accounts, false)
		if err != nil {
			t.Fatal(err)
		}
		clean, err := overview(portfolio("10000"), m, "RUB", from, today, accounts, false)
		if err != nil {
			t.Fatal(err)
		}
		table, err := positionTable(withGap, m, "RUB", from, today, accounts, false)
		if err != nil {
			t.Fatal(err)
		}

		// накануне 9000 + 10 × 120 = 10200, сегодня 11400, вложено 1000: прибыль 200 — с расхождением и без.
		if card.Profit != "200.00" || clean.Profit != "200.00" {
			t.Errorf("profit with gap %s, without %s; want 200.00 both", card.Profit, clean.Profit)
		}
		// стоимость в карточке всё равно как у брокера.
		if card.Value != "11300.00" {
			t.Errorf("card value = %s, want 11300.00", card.Value)
		}
		// строка расхождения есть: в стоимости −100, в прибыли пусто; итоги — как в карточках.
		if table.Unexplained != "-100.00" || table.UnexplainedProfit != "" {
			t.Errorf("unexplained = %q / %q, want -100.00 and empty", table.Unexplained, table.UnexplainedProfit)
		}
		if table.TotalValue != card.Value || table.TotalProfit != card.Profit {
			t.Errorf("table totals %s / %s, card %s / %s", table.TotalValue, table.TotalProfit, card.Value, card.Profit)
		}
	})
}
