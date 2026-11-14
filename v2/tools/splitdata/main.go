package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
)

const brokerAPI = "https://invest-public-api.tbank.ru/rest/tinkoff.public.invest.api.contract.v1."

type moexSplit struct {
	TradeDate string `json:"tradedate"`
	SecID     string `json:"secid"`
	Before    int64  `json:"before"`
	After     int64  `json:"after"`
}

type share struct {
	Ticker    string `json:"ticker"`
	ClassCode string `json:"classCode"`
	ISIN      string `json:"isin"`
	AssetUID  string `json:"assetUid"`
	Name      string `json:"name"`
}

type yahooSplit struct {
	Ticker      string  `json:"ticker"`
	ClassCode   string  `json:"classCode"`
	ISIN        string  `json:"isin"`
	AssetUID    string  `json:"assetUid"`
	Name        string  `json:"name"`
	Symbol      string  `json:"yahooSymbol"`
	Date        string  `json:"date"`
	Numerator   float64 `json:"numerator"`
	Denominator float64 `json:"denominator"`
}

var isinLike = regexp.MustCompile(`^[A-Z]{2}[A-Z0-9]{9}[0-9]$`)

func main() {
	_ = godotenv.Load()
	token := os.Getenv("T_INVEST_TOKEN_ALL")
	if token == "" {
		fmt.Println("T_INVEST_TOKEN_ALL is not set")
		os.Exit(1)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	moex, err := fetchMOEX(ctx)
	if err != nil {
		fmt.Println("moex:", err)
	} else {
		save("splits_moex.json", moex)
		fmt.Printf("Мосбиржа: %d записей, с %s по %s\n", len(moex), moex[0].TradeDate, moex[len(moex)-1].TradeDate)
	}

	assets, order, err := foreignAssets(ctx, token)
	if err != nil {
		fmt.Println("broker:", err)
		os.Exit(1)
	}
	fmt.Printf("иностранных активов СПБ: %d\n", len(order))

	var mu sync.Mutex
	var done atomic.Int64
	var found []yahooSplit
	var missed []string

	var g errgroup.Group
	g.SetLimit(100)
	for _, assetUID := range order {
		g.Go(func() error {
			if ctx.Err() != nil {
				return nil
			}
			group := assets[assetUID]
			var local []yahooSplit
			var lastErr error
			ok := false
			for _, c := range candidates(group) {
				splits, err := fetchYahoo(ctx, c.symbol)
				time.Sleep(100 * time.Millisecond)
				if err != nil {
					lastErr = err
					continue
				}
				for _, sp := range splits {
					sp.Ticker, sp.ClassCode, sp.ISIN, sp.AssetUID, sp.Name, sp.Symbol =
						c.share.Ticker, c.share.ClassCode, c.share.ISIN, assetUID, c.share.Name, c.symbol
					local = append(local, sp)
				}
				ok = true
				break
			}

			mu.Lock()
			found = append(found, local...)
			if !ok {
				var tickers []string
				for _, s := range group {
					tickers = append(tickers, s.Ticker+"/"+s.ClassCode)
				}
				missed = append(missed, fmt.Sprintf("%s\t%s\t%v", assetUID, strings.Join(tickers, ","), lastErr))
			}
			n := done.Add(1)
			fmt.Fprintf(os.Stderr, "\r%d/%d  найдено сплитов: %d  без данных: %d\033[K", n, len(order), len(found), len(missed))
			mu.Unlock()
			return nil
		})
	}
	_ = g.Wait()
	fmt.Fprintln(os.Stderr)

	sort.Slice(found, func(a, b int) bool { return found[a].Date < found[b].Date })
	save("splits_yahoo.json", found)
	_ = os.WriteFile("splits_yahoo_missed.txt", []byte(strings.Join(missed, "\n")+"\n"), 0o644)
	fmt.Printf("Yahoo: %d сплитов, без данных %d активов\n", len(found), len(missed))
}

type candidate struct {
	symbol string
	share  share
}

func candidates(group []share) []candidate {
	var result, fallback []candidate
	seen := make(map[string]bool)
	add := func(list *[]candidate, symbol string, s share) {
		if symbol == "" || seen[symbol] {
			return
		}
		seen[symbol] = true
		*list = append(*list, candidate{symbol, s})
	}
	for _, s := range group {
		add(&result, yahooSymbol(s.Ticker, s.ClassCode), s)
	}
	for _, s := range group {
		t := strings.TrimSuffix(strings.TrimSuffix(s.Ticker, "_old"), "old")
		if t != s.Ticker {
			add(&fallback, yahooSymbol(t, s.ClassCode), s)
		}
	}
	return append(result, fallback...)
}

func yahooSymbol(ticker, class string) string {
	if isinLike.MatchString(ticker) {
		return ""
	}
	if base, suffix, ok := strings.Cut(ticker, "@"); ok {
		if suffix != "DE" {
			return ""
		}
		return strings.ReplaceAll(base, ".", "-") + ".DE"
	}
	t := strings.ReplaceAll(ticker, ".", "-")
	switch {
	case class == "SPBHKEX":
		t = strings.TrimLeft(t, "0")
		for len(t) < 4 {
			t = "0" + t
		}
		return t + ".HK"
	case strings.HasPrefix(class, "SPBDE"):
		return t + ".DE"
	}
	return t
}

func fetchMOEX(ctx context.Context) ([]moexSplit, error) {
	var all []moexSplit
	seen := make(map[string]bool)
	for start := 0; ; start += 100 {
		url := fmt.Sprintf("https://iss.moex.com/iss/statistics/engines/stock/splits.json?from=1990-01-01&start=%d", start)
		body, err := get(ctx, url)
		if err != nil {
			return nil, err
		}
		var out struct {
			Splits struct {
				Data [][]json.RawMessage `json:"data"`
			} `json:"splits"`
		}
		if err := json.Unmarshal(body, &out); err != nil {
			return nil, err
		}
		added := 0
		for _, row := range out.Splits.Data {
			if len(row) < 4 {
				continue
			}
			var s moexSplit
			_ = json.Unmarshal(row[0], &s.TradeDate)
			_ = json.Unmarshal(row[1], &s.SecID)
			_ = json.Unmarshal(row[2], &s.Before)
			_ = json.Unmarshal(row[3], &s.After)
			key := s.TradeDate + s.SecID
			if seen[key] {
				continue
			}
			seen[key] = true
			all = append(all, s)
			added++
		}
		if added == 0 {
			break
		}
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("empty response")
	}
	sort.Slice(all, func(a, b int) bool { return all[a].TradeDate < all[b].TradeDate })
	return all, nil
}

func foreignAssets(ctx context.Context, token string) (map[string][]share, []string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, brokerAPI+"InstrumentsService/Shares",
		strings.NewReader(`{"instrumentStatus":"INSTRUMENT_STATUS_ALL"}`))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var out struct {
		Instruments []share `json:"instruments"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, nil, err
	}

	assets := make(map[string][]share)
	for _, s := range out.Instruments {
		if !strings.HasPrefix(s.ClassCode, "SPB") || s.AssetUID == "" {
			continue
		}
		assets[s.AssetUID] = append(assets[s.AssetUID], s)
	}
	order := make([]string, 0, len(assets))
	for uid := range assets {
		order = append(order, uid)
	}
	sort.Slice(order, func(a, b int) bool { return assets[order[a]][0].Ticker < assets[order[b]][0].Ticker })
	return assets, order, nil
}

func fetchYahoo(ctx context.Context, symbol string) ([]yahooSplit, error) {
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?period1=0&period2=%d&interval=1mo&events=split",
		symbol, time.Now().Unix())

	var body []byte
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		body, err = get(ctx, url)
		if err == nil || !strings.Contains(err.Error(), "429") {
			break
		}
		time.Sleep(time.Duration(attempt+1) * 5 * time.Second)
	}
	if err != nil {
		return nil, err
	}

	var out struct {
		Chart struct {
			Result []struct {
				Events struct {
					Splits map[string]struct {
						Date        int64   `json:"date"`
						Numerator   float64 `json:"numerator"`
						Denominator float64 `json:"denominator"`
					} `json:"splits"`
				} `json:"events"`
			} `json:"result"`
		} `json:"chart"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	var result []yahooSplit
	for _, r := range out.Chart.Result {
		for _, s := range r.Events.Splits {
			result = append(result, yahooSplit{
				Date:        time.Unix(s.Date, 0).UTC().Format("2006-01-02"),
				Numerator:   s.Numerator,
				Denominator: s.Denominator,
			})
		}
	}
	return result, nil
}

func get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return body, nil
}

func save(name string, v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	_ = os.WriteFile(name, b, 0o644)
}
