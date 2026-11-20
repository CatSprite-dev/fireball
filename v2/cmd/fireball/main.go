package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"slices"
	"time"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"

	"github.com/CatSprite-dev/fireball/v2/internal/convert"
	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/load"
	"github.com/CatSprite-dev/fireball/v2/internal/metrics"
	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/CatSprite-dev/fireball/v2/internal/snapshot"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
)

const baseCurrency = "RUB"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if err := run(logger); err != nil {
		logger.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return fmt.Errorf("load .env: %w", err)
		}
	}

	token := os.Getenv("T_INVEST_TOKEN_ALL")
	if token == "" {
		return errors.New("T_INVEST_TOKEN_ALL environment variable is not set")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client := tinvest.New(token, logger)

	loader := load.New(client, baseCurrency, logger)

	p, err := loader.Portfolio(ctx)
	if err != nil {
		return err
	}

	to := series.Day(time.Now())
	from := to.AddDate(-1, 0, 0)
	if slices.Contains(os.Args[1:], "all") {
		from = series.Day(p.Events[0].Date)
	}
	loadFrom := from.AddDate(0, 0, -30)

	m, err := loader.Market(ctx, p, loadFrom, time.Now())
	if err != nil {
		return err
	}
	index, err := loader.Index(ctx, "MCFTR", loadFrom, time.Now())
	if err != nil {
		return err
	}

	if err := checkTWR(p, m, index, nil, from, to); err != nil {
		return err
	}
	return checkAccounts(p)
}

func checkAccounts(p *load.Portfolio) error {
	events, cur := p.Events, p.Currencies
	var sumTrade, sumPayout, sumClearing, sumAccount, sumBroker decimal.Decimal
	fmt.Printf("%-28s %12s %12s %10s %10s %12s %12s %8s\n",
		"счёт", "сделки", "выплаты", "клиринг", "счёт", "итого", "брокер", "разница")

	for _, d := range p.Accounts {
		evs := ledger.Filter(events, []string{d.Account.ID})
		book := ledger.ByInstrument(evs, cur.ByAsset)
		keys := ledger.PageKeys(evs, cur.ByAsset)

		value := make(map[string]decimal.Decimal)
		qty := make(map[string]decimal.Decimal)
		if d.Portfolio != nil {
			holdings, err := convert.ToHoldings(d.Account.ID, d.Portfolio)
			if err != nil {
				return fmt.Errorf("holdings %s: %w", d.Account.Name, err)
			}
			for _, h := range holdings {
				key, ok := keys[h.InstrumentUID]
				if code, isCur := cur.ByUID[h.InstrumentUID]; isCur {
					if code == baseCurrency {
						continue
					}
					key, ok = code, true
				}
				if !ok {
					key = h.InstrumentUID
				}
				value[key] = value[key].Add(h.Value())
				qty[key] = qty[key].Add(h.Quantity)
			}
		}

		var trade, payout, clearing, account decimal.Decimal
		for key, f := range book {
			trade = trade.Add(value[key]).Add(f.Bought).Add(f.Sold).Add(f.Fees)
			payout = payout.Add(f.Income).Add(f.Taxes)
			clearing = clearing.Add(f.Clearing)
			if !f.Quantity.Equal(qty[key]) {
				fmt.Printf("  %s: книга %s, брокер %s\n", key, f.Quantity, qty[key])
			}
		}
		for key, q := range qty {
			if _, ok := book[key]; !ok {
				fmt.Printf("  %s: нет в книге, брокер %s\n", key, q)
			}
		}
		for _, e := range evs {
			noPaper := e.InstrumentUID == "" && e.AssetUID == ""
			if noPaper && (e.OperationKind == ledger.KindFee || e.OperationKind == ledger.KindTax) {
				account = account.Add(e.BaseAmount.Amount)
			}
		}

		brokerValue := decimal.Zero
		if d.Portfolio != nil {
			mv := d.Portfolio.TotalAmountPortfolio
			brokerValue = decimal.RequireFromString(mv.Units).Add(decimal.New(int64(mv.Nano), -9))
		}
		netIn := decimal.Zero
		for _, e := range evs {
			switch e.OperationKind {
			case ledger.KindDeposit, ledger.KindWithdrawal, ledger.KindTransferIn, ledger.KindTransferOut:
				netIn = netIn.Add(e.BaseAmount.Amount)
			}
		}
		broker := brokerValue.Sub(netIn)

		sumTrade, sumPayout = sumTrade.Add(trade), sumPayout.Add(payout)
		sumClearing, sumAccount = sumClearing.Add(clearing), sumAccount.Add(account)
		sumBroker = sumBroker.Add(broker)
		printSplit(d.Account.Name, trade, payout, clearing, account, broker)
	}
	printSplit("ИТОГО", sumTrade, sumPayout, sumClearing, sumAccount, sumBroker)
	return nil
}

func printSplit(name string, trade, payout, clearing, account, broker decimal.Decimal) {
	total := trade.Add(payout).Add(clearing).Add(account)
	fmt.Printf("%-28s %12s %12s %10s %10s %12s %12s %8s\n", name,
		trade.StringFixed(2), payout.StringFixed(2), clearing.StringFixed(2), account.StringFixed(2),
		total.StringFixed(2), broker.StringFixed(2), total.Sub(broker).StringFixed(2))
}

func checkTWR(p *load.Portfolio, m *load.Market, index series.Series, accountIDs []string, from, to time.Time) error {
	events := p.Events
	if accountIDs != nil {
		events = ledger.Filter(events, accountIDs)
	}
	positions := ledger.ByInstrument(events, p.Currencies.ByAsset)

	cash := ledger.CashHistory(events, baseCurrency)
	rows, gaps := snapshot.Build(positions, m.Prices, m.Bonds, cash, baseCurrency, from, to)
	values := snapshot.Total(rows, from, to)
	deposits := ledger.NetDeposits(events)
	in, out := ledger.DepositsAndWithdrawals(events)
	growth := metrics.TWR(values, in, out, from, to)

	fmt.Printf("период: %s — %s\n", from.Format(time.DateOnly), to.Format(time.DateOnly))
	fmt.Printf("дыр: %d\n", len(gaps))
	for _, g := range gaps {
		fmt.Printf("  дыра %s %s — %s\n", g.Key, g.From.Format(time.DateOnly), g.To.Format(time.DateOnly))
	}

	hundred := decimal.NewFromInt(100)
	twr := growth[len(growth)-1].Value.Sub(decimal.NewFromInt(1)).Mul(hundred)
	fmt.Printf("TWR за период: %s%%\n", twr.StringFixed(2))
	r, err := metrics.MWR(values, deposits, from, to)
	if err != nil {
		return err
	}
	periodDays := to.Sub(from).Hours() / 24
	period := math.Pow(1+r, periodDays/365) - 1
	fmt.Printf("MWR: %.2f%% годовых, за период %.2f%%\n", r*100, period*100)

	sumDeposits := func(after, upTo time.Time) decimal.Decimal {
		s := decimal.Zero
		for _, pt := range deposits {
			if pt.Date.After(after) && !pt.Date.After(upTo) {
				s = s.Add(pt.Value)
			}
		}
		return s
	}
	vFrom, vTo := values[0].Value, values[len(values)-1].Value
	fmt.Printf("прибыль за период: %s ₽\n", vTo.Sub(vFrom).Sub(sumDeposits(from, to)).StringFixed(2))
	fmt.Printf("прибыль за всё время (по снимку): %s ₽\n", vTo.Sub(sumDeposits(time.Time{}, to)).StringFixed(2))

	cmpFrom := from
	if first := index[0].Date; first.After(cmpFrom) {
		cmpFrom = first
		fmt.Printf("индекс есть только с %s — сравнение с индексом с этой даты\n", cmpFrom.Format(time.DateOnly))
	}
	start, ok1 := index.At(cmpFrom)
	end, ok2 := index.At(to)
	g0, ok3 := growth.At(cmpFrom)
	if !ok1 || !ok2 || !ok3 {
		return errors.New("MCFTR: no value at comparison bounds")
	}
	indexReturn := end.Value.Div(start.Value).Sub(decimal.NewFromInt(1)).Mul(hundred)
	twrCmp := growth[len(growth)-1].Value.Div(g0.Value).Sub(decimal.NewFromInt(1)).Mul(hundred)
	fmt.Printf("с %s: TWR %s%%, MCFTR %s%% (%s → %s), разница %s п.п.\n",
		cmpFrom.Format(time.DateOnly), twrCmp.StringFixed(2), indexReturn.StringFixed(2),
		start.Value.StringFixed(2), end.Value.StringFixed(2), twrCmp.Sub(indexReturn).StringFixed(2))

	var secRows []snapshot.Row
	for _, r := range rows {
		switch r.Kind {
		case ledger.InstrumentShare, ledger.InstrumentBond, ledger.InstrumentETF:
			secRows = append(secRows, r)
		}
	}
	securities := snapshot.Total(secRows, from, to)
	trades, income := ledger.SecurityFlows(events)
	actual, shadow, err := metrics.ShadowTrades(securities, index, trades, income, cmpFrom, to)
	if err != nil {
		return err
	}
	fmt.Println("по сделкам: ты / тень MCFTR / разница")
	for i, pt := range actual {
		if (pt.Date.Day() == 1 && pt.Date.Month() == time.January) || pt.Date.Equal(to) {
			fmt.Printf("  %s %14s %14s %12s\n", pt.Date.Format(time.DateOnly),
				pt.Value.StringFixed(2), shadow[i].Value.StringFixed(2), pt.Value.Sub(shadow[i].Value).StringFixed(2))
		}
	}

	shadowDep, err := metrics.ShadowDeposits(values, index, deposits, cmpFrom, to)
	if err != nil {
		return err
	}
	fmt.Println("по пополнениям: ты / тень MCFTR / разница")
	for _, pt := range shadowDep {
		v, _ := values.At(pt.Date)
		if (pt.Date.Day() == 1 && pt.Date.Month() == time.January) || pt.Date.Equal(to) {
			fmt.Printf("  %s %14s %14s %12s\n", pt.Date.Format(time.DateOnly),
				v.Value.StringFixed(2), pt.Value.StringFixed(2), v.Value.Sub(pt.Value).StringFixed(2))
		}
	}

	if slices.Contains(os.Args[1:], "csv") {
		cols := map[string]series.Series{
			"value": values, "deposit": deposits, "twr": growth,
			"index":           metrics.TWR(index, nil, nil, from, to),
			"shadow_deposits": shadowDep, "securities": securities, "shadow_trades": shadow,
		}
		if err := writeSeries("series.csv", values, cols); err != nil {
			return err
		}
	}

	return checkValues(p, events, rows, to)
}

func writeSeries(path string, days series.Series, cols map[string]series.Series) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	names := []string{"value", "deposit", "twr", "index", "shadow_deposits", "securities", "shadow_trades"}
	w := csv.NewWriter(f)
	w.Write(append([]string{"date"}, names...))
	for _, d := range days {
		row := []string{d.Date.Format(time.DateOnly)}
		for _, n := range names {
			cell := ""
			if pt, ok := cols[n].At(d.Date); ok && pt.Date.Equal(d.Date) {
				cell = pt.Value.StringFixed(6)
			}
			row = append(row, cell)
		}
		w.Write(row)
	}
	w.Flush()
	return w.Error()
}

func checkValues(p *load.Portfolio, events []ledger.Event, rows []snapshot.Row, day time.Time) error {
	cur := p.Currencies
	keys := ledger.PageKeys(events, cur.ByAsset)

	type pair struct {
		ours, broker           decimal.Decimal
		qty, ourPrice, brPrice decimal.Decimal
		priceDate              time.Time
	}
	byKey := make(map[string]*pair)
	get := func(k string) *pair {
		if byKey[k] == nil {
			byKey[k] = &pair{}
		}
		return byKey[k]
	}

	for _, r := range rows {
		if !r.Date.Equal(day) {
			continue
		}
		pr := get(r.Key)
		pr.ours = pr.ours.Add(r.Value)
		pr.qty = pr.qty.Add(r.Quantity)
		pr.ourPrice = r.Price.Value.Add(r.NKD)
		pr.priceDate = r.Price.Date
	}
	for _, d := range p.Accounts {
		if d.Portfolio == nil {
			continue
		}
		holdings, err := convert.ToHoldings(d.Account.ID, d.Portfolio)
		if err != nil {
			return fmt.Errorf("holdings %s: %w", d.Account.Name, err)
		}
		for _, h := range holdings {
			key, ok := keys[h.InstrumentUID]
			if code, isCur := cur.ByUID[h.InstrumentUID]; isCur {
				key, ok = code, true
			}
			if !ok {
				key = h.InstrumentUID
			}
			pr := get(key)
			pr.broker = pr.broker.Add(h.Value())
			pr.brPrice = h.Price.Add(h.NKD)
		}
	}

	total := decimal.Zero
	fmt.Println("стоимость по позициям (наша − брокер), где разница больше 1 ₽:")
	for k, pr := range byKey {
		diff := pr.ours.Sub(pr.broker)
		total = total.Add(diff)
		if diff.Abs().LessThan(decimal.NewFromInt(1)) {
			continue
		}
		fmt.Printf("  %-38s qty=%-8s наша цена=%-10s (от %s) брокер=%-10s  разница=%s\n",
			k, pr.qty.String(), pr.ourPrice.StringFixed(2), pr.priceDate.Format(time.DateOnly),
			pr.brPrice.StringFixed(2), diff.StringFixed(2))
	}
	fmt.Printf("  итого разница: %s\n", total.StringFixed(2))
	return nil
}

func dumpEvents(events []ledger.Event, from, to time.Time) {
	for _, e := range events {
		if e.Date.Before(from) || e.Date.After(to) {
			continue
		}
		fmt.Printf("%s  %-8s %-16s %-10s qty=%-6s %12s %s\n",
			e.Date.Format("2006-01-02 15:04"), e.AccountID[:8], e.OperationKind, e.InstrumentKind,
			e.Quantity.String(), e.Amount.Amount.StringFixed(2), e.Amount.Currency)
	}
}
