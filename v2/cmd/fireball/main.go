package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/signal"
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
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

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

	loader := load.New(client, baseCurrency)

	start := time.Now()
	p, err := loader.Portfolio(ctx)
	if err != nil {
		return err
	}
	logger.Info("loaded", "accounts", len(p.Accounts), "events", len(p.Events), "duration", time.Since(start))

	to := series.Day(time.Now())
	from := to.AddDate(-1, 0, 0)
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
	growth := metrics.TWR(values, deposits, from, to)

	fmt.Printf("дыр: %d\n", len(gaps))
	for _, g := range gaps {
		fmt.Printf("  дыра %s %s — %s\n", g.Key, g.From.Format(time.DateOnly), g.To.Format(time.DateOnly))
	}

	twr := growth[len(growth)-1].Value.Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100))
	fmt.Printf("TWR за 12 месяцев: %s%%\n", twr.StringFixed(2))
	r, err := metrics.MWR(values, deposits, from, to)
	if err != nil {
		return err
	}
	periodDays := to.Sub(from).Hours() / 24
	period := math.Pow(1+r, periodDays/365) - 1
	fmt.Printf("MWR: %.2f%% годовых, за период %.2f%%\n", r*100, period*100)

	sumDeposits := func(after, upTo time.Time) decimal.Decimal {
		s := decimal.Zero
		for _, p := range deposits {
			if p.Date.After(after) && !p.Date.After(upTo) {
				s = s.Add(p.Value)
			}
		}
		return s
	}
	vFrom, vTo := values[0].Value, values[len(values)-1].Value
	periodProfit := vTo.Sub(vFrom).Sub(sumDeposits(from, to))
	allTimeProfit := vTo.Sub(sumDeposits(time.Time{}, to))
	fmt.Printf("прибыль за 12 месяцев: %s ₽\n", periodProfit.StringFixed(2))
	fmt.Printf("прибыль за всё время (по снимку): %s ₽\n", allTimeProfit.StringFixed(2))
	start, ok1 := index.At(from)
	end, ok2 := index.At(to)
	if !ok1 || !ok2 {
		return errors.New("MCFTR: no value at period bounds")
	}
	indexReturn := end.Value.Div(start.Value).Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100))
	fmt.Printf("MCFTR %s → %s: %s → %s, %s%%\n",
		start.Date.Format(time.DateOnly), end.Date.Format(time.DateOnly),
		start.Value.StringFixed(2), end.Value.StringFixed(2), indexReturn.StringFixed(2))
	fmt.Printf("TWR против MCFTR: %s п.п.\n", twr.Sub(indexReturn).StringFixed(2))

	var secRows []snapshot.Row
	for _, r := range rows {
		switch r.Kind {
		case ledger.InstrumentShare, ledger.InstrumentBond, ledger.InstrumentETF:
			secRows = append(secRows, r)
		}
	}
	securities := snapshot.Total(secRows, from, to)
	trades, income := ledger.SecurityFlows(events)
	actual, shadow, err := metrics.ShadowTrades(securities, index, trades, income, from, to)
	if err != nil {
		return err
	}
	fmt.Println("по сделкам: ты / тень MCFTR / разница")
	for i, p := range actual {
		if p.Date.Day() == 1 || p.Date.Equal(to) {
			fmt.Printf("  %s %14s %14s %12s\n", p.Date.Format(time.DateOnly),
				p.Value.StringFixed(2), shadow[i].Value.StringFixed(2), p.Value.Sub(shadow[i].Value).StringFixed(2))
		}
	}

	shadowDep, err := metrics.ShadowDeposits(values, index, deposits, from, to)
	if err != nil {
		return err
	}
	fmt.Println("по пополнениям: ты / тень MCFTR / разница")
	for i, p := range values {
		if p.Date.Day() == 1 || p.Date.Equal(to) {
			fmt.Printf("  %s %14s %14s %12s\n", p.Date.Format(time.DateOnly),
				p.Value.StringFixed(2), shadowDep[i].Value.StringFixed(2), p.Value.Sub(shadowDep[i].Value).StringFixed(2))
		}
	}
	if err := checkValues(p, events, rows, to); err != nil {
		return err
	}

	return nil
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
		p := get(r.Key)
		p.ours = p.ours.Add(r.Value)
		p.qty = p.qty.Add(r.Quantity)
		p.ourPrice = r.Price.Value.Add(r.NKD)
		p.priceDate = r.Price.Date
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
			p := get(key)
			p.broker = p.broker.Add(h.Value())
			p.brPrice = h.Price.Add(h.NKD)
		}
	}

	total := decimal.Zero
	fmt.Println("стоимость по позициям (наша − брокер), где разница больше 1 ₽:")
	for k, p := range byKey {
		diff := p.ours.Sub(p.broker)
		total = total.Add(diff)
		if diff.Abs().LessThan(decimal.NewFromInt(1)) {
			continue
		}
		fmt.Printf("  %-38s qty=%-8s наша цена=%-10s (от %s) брокер=%-10s  разница=%s\n",
			k, p.qty.String(), p.ourPrice.StringFixed(2), p.priceDate.Format(time.DateOnly),
			p.brPrice.StringFixed(2), diff.StringFixed(2))
	}
	fmt.Printf("  итого разница: %s\n", total.StringFixed(2))
	return nil
}
