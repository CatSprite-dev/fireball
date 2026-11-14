package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/format"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const brokerAPI = "https://invest-public-api.tbank.ru/rest/tinkoff.public.invest.api.contract.v1."

const (
	since          = "2014-01-01"
	maxSmallerSide = 10
	maxDateGap     = 7
)

type moexSplit struct {
	TradeDate string `json:"tradedate"`
	SecID     string `json:"secid"`
	Before    int64  `json:"before"`
	After     int64  `json:"after"`
}

type yahooSplit struct {
	Ticker      string  `json:"ticker"`
	AssetUID    string  `json:"assetUid"`
	Symbol      string  `json:"yahooSymbol"`
	Date        string  `json:"date"`
	Numerator   float64 `json:"numerator"`
	Denominator float64 `json:"denominator"`
}

type instrument struct {
	Ticker    string `json:"ticker"`
	ClassCode string `json:"classCode"`
	AssetUID  string `json:"assetUid"`
	Name      string `json:"name"`
}

type split struct {
	assetUID string
	date     string
	before   int64
	after    int64
	source   string
}

func main() {
	_ = godotenv.Load()
	token := os.Getenv("T_INVEST_TOKEN_ALL")
	if token == "" {
		fmt.Println("T_INVEST_TOKEN_ALL is not set")
		os.Exit(1)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var moex []moexSplit
	if err := load("data/splits/splits_moex.json", &moex); err != nil {
		fmt.Println("moex:", err)
		os.Exit(1)
	}

	instruments, err := brokerInstruments(ctx, token)
	if err != nil {
		fmt.Println("broker:", err)
		os.Exit(1)
	}

	instrumentsByTicker := make(map[string][]instrument)
	for _, in := range instruments {
		instrumentsByTicker[in.Ticker] = append(instrumentsByTicker[in.Ticker], in)
	}

	var moexSplits []split
	var unmatched []string
	for _, s := range moex {
		matches := instrumentsByTicker[s.SecID]
		if len(matches) == 0 {
			unmatched = append(unmatched, s.SecID)
			continue
		}
		assetUID, ok := moexAssetUID(matches)
		if !ok {
			fmt.Printf("конфликт: %s %s — несколько активов, пропущено\n", s.TradeDate, s.SecID)
			continue
		}
		moexSplits = append(moexSplits, split{assetUID: assetUID, date: s.TradeDate, before: s.Before, after: s.After, source: "moex " + s.SecID})
	}
	fmt.Printf("нет у брокера: %s\n", strings.Join(unmatched, " "))

	var yahoo []yahooSplit
	if err := load("data/splits/splits_yahoo.json", &yahoo); err != nil {
		fmt.Println("yahoo:", err)
		os.Exit(1)
	}
	yahooSplits := fromYahoo(yahoo)

	splits := merge(yahooSplits, moexSplits)
	fmt.Printf("итого: %d (yahoo %d, moex %d, дублей %d)\n",
		len(splits), len(yahooSplits), len(moexSplits), len(yahooSplits)+len(moexSplits)-len(splits))

	if err := writeGo("internal/ledger/splits_generated.go", splits); err != nil {
		fmt.Println("write:", err)
		os.Exit(1)
	}
	fmt.Println("записан internal/ledger/splits_generated.go")
}

func moexAssetUID(matches []instrument) (string, bool) {
	all := make(map[string]bool)
	onMainBoard := make(map[string]bool)
	for _, m := range matches {
		all[m.AssetUID] = true
		if strings.HasPrefix(m.ClassCode, "TQ") {
			onMainBoard[m.AssetUID] = true
		}
	}
	if uid, ok := single(all); ok {
		return uid, true
	}
	return single(onMainBoard)
}

func single(set map[string]bool) (string, bool) {
	if len(set) != 1 {
		return "", false
	}
	for uid := range set {
		return uid, true
	}
	return "", false
}

func fromYahoo(rows []yahooSplit) []split {
	var result []split
	seen := make(map[string]bool)
	for _, r := range rows {
		if r.Date < since {
			continue
		}
		before, after, ok := ratio(r.Denominator, r.Numerator)
		if !ok {
			fmt.Printf("выброшен: %s  %-8s %g:%g  %s\n", r.Date, r.Ticker, r.Denominator, r.Numerator, r.AssetUID)
			continue
		}
		key := fmt.Sprintf("%s %s %d:%d", r.AssetUID, r.Date, before, after)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, split{assetUID: r.AssetUID, date: r.Date, before: before, after: after, source: "yahoo " + r.Symbol})
	}
	return result
}

func ratio(before, after float64) (int64, int64, bool) {
	b, a := int64(before), int64(after)
	if float64(b) != before || float64(a) != after || b <= 0 || a <= 0 {
		return 0, 0, false
	}
	g := gcd(b, a)
	b, a = b/g, a/g
	if b == a || min(b, a) > maxSmallerSide {
		return 0, 0, false
	}
	return b, a, true
}

func gcd(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func merge(yahoo, moex []split) []split {
	result := append([]split(nil), yahoo...)
	for _, m := range moex {
		if dup, ok := findSame(yahoo, m); ok {
			fmt.Printf("дубль: %s %s → берём %s %s\n", m.source, m.date, dup.source, dup.date)
			continue
		}
		result = append(result, m)
	}
	return result
}

func findSame(list []split, s split) (split, bool) {
	for _, x := range list {
		if x.assetUID == s.assetUID && x.before == s.before && x.after == s.after &&
			daysApart(x.date, s.date) <= maxDateGap {
			return x, true
		}
	}
	return split{}, false
}

func daysApart(a, b string) int {
	d := int(mustDate(a).Sub(mustDate(b)).Hours() / 24)
	if d < 0 {
		d = -d
	}
	return d
}

func mustDate(s string) time.Time {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}
	return t
}

func brokerInstruments(ctx context.Context, token string) ([]instrument, error) {
	var all []instrument
	for _, method := range []string{"Shares", "Etfs"} {
		list, err := fetchInstruments(ctx, token, method)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", method, err)
		}
		all = append(all, list...)
	}
	return all, nil
}

func fetchInstruments(ctx context.Context, token, method string) ([]instrument, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, brokerAPI+"InstrumentsService/"+method,
		strings.NewReader(`{"instrumentStatus":"INSTRUMENT_STATUS_ALL"}`))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var out struct {
		Instruments []instrument `json:"instruments"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Instruments, nil
}

func load(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func writeGo(path string, splits []split) error {
	splitsByAsset := make(map[string][]split)
	for _, s := range splits {
		splitsByAsset[s.assetUID] = append(splitsByAsset[s.assetUID], s)
	}
	uids := make([]string, 0, len(splitsByAsset))
	for uid := range splitsByAsset {
		uids = append(uids, uid)
	}
	sort.Strings(uids)

	var b bytes.Buffer
	b.WriteString("// Code generated by splitgen. DO NOT EDIT.\n\npackage ledger\n\nimport \"time\"\n\n")
	b.WriteString("var generatedSplitsByAssetUID = map[string][]split{\n")
	for _, uid := range uids {
		list := splitsByAsset[uid]
		sort.Slice(list, func(i, j int) bool { return list[i].date < list[j].date })
		fmt.Fprintf(&b, "%q: {\n", uid)
		for _, s := range list {
			d := mustDate(s.date)
			fmt.Fprintf(&b, "{date: time.Date(%d, %d, %d, 0, 0, 0, 0, moscow), before: %d, after: %d}, // %s\n",
				d.Year(), d.Month(), d.Day(), s.before, s.after, s.source)
		}
		b.WriteString("},\n")
	}
	b.WriteString("}\n")

	src, err := format.Source(b.Bytes())
	if err != nil {
		return err
	}
	return os.WriteFile(path, src, 0o644)
}
