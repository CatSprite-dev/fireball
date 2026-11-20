package metrics

import (
	"testing"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/shopspring/decimal"
)

func TestShadowTrades(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, time.March, d, 0, 0, 0, 0, time.UTC) }
	pt := func(d int, v string) series.Point {
		return series.Point{Date: day(d), Value: decimal.RequireFromString(v)}
	}
	near := func(got decimal.Decimal, want string) bool {
		return got.Sub(decimal.RequireFromString(want)).Abs().LessThan(decimal.RequireFromString("0.01"))
	}

	cases := []struct {
		name                      string
		securities, index, trades series.Series
		income                    series.Series
		from, to                  int
		real, shadow              []string // на каждый день от from до to
	}{
		{
			name:       "купон реинвестирован",
			securities: series.Series{pt(1, "1000"), pt(2, "1050"), pt(3, "1000"), pt(4, "1050"), pt(5, "1080")},
			index:      series.Series{pt(1, "100"), pt(2, "105"), pt(3, "105"), pt(4, "105"), pt(5, "110")},
			trades:     series.Series{pt(4, "-50")},
			income:     series.Series{pt(3, "50")},
			from:       1, to: 5,
			real:   []string{"1000", "1050", "1000", "1050", "1080"},
			shadow: []string{"1000", "1050", "1000", "1050", "1100"},
		},
		{
			name:       "купон не реинвестирован",
			securities: series.Series{pt(1, "1000"), pt(2, "1050"), pt(3, "1000"), pt(5, "1030")},
			index:      series.Series{pt(1, "100"), pt(2, "105"), pt(3, "105"), pt(5, "110")},
			income:     series.Series{pt(3, "50")},
			from:       1, to: 5,
			real:   []string{"1000", "1050", "1000", "1000", "1030"},
			shadow: []string{"1000", "1050", "1000", "1000", "1047.62"},
		},
		{
			name:       "комиссия — тень докупает",
			securities: series.Series{pt(1, "0"), pt(2, "1000")},
			index:      series.Series{pt(1, "100"), pt(2, "100")},
			trades:     series.Series{pt(2, "-1000")},
			income:     series.Series{pt(2, "-3")},
			from:       1, to: 2,
			real:   []string{"0", "1000"},
			shadow: []string{"0", "1003"},
		},
		{
			name:       "продажа дороже индекса — паи в минусе",
			securities: series.Series{pt(1, "10000"), pt(2, "0")},
			index:      series.Series{pt(1, "1000"), pt(2, "1100")},
			trades:     series.Series{pt(2, "12000")},
			from:       1, to: 2,
			real:   []string{"10000", "0"},
			shadow: []string{"10000", "-1000"},
		},
	}
	for _, c := range cases {
		real, shadow, err := ShadowTrades(c.securities, c.index, c.trades, c.income, day(c.from), day(c.to))
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if len(real) != len(c.real) || len(shadow) != len(c.shadow) {
			t.Errorf("%s: got %d real and %d shadow points, want %d", c.name, len(real), len(shadow), len(c.real))
			continue
		}
		for i := range c.real {
			if !real[i].Date.Equal(day(c.from+i)) || !shadow[i].Date.Equal(day(c.from+i)) {
				t.Errorf("%s: point %d: dates %s and %s, want %s", c.name, i,
					real[i].Date.Format(time.DateOnly), shadow[i].Date.Format(time.DateOnly), day(c.from+i).Format(time.DateOnly))
			}
			if !near(real[i].Value, c.real[i]) || !near(shadow[i].Value, c.shadow[i]) {
				t.Errorf("%s: point %d: real %s shadow %s, want %s and %s",
					c.name, i, real[i].Value, shadow[i].Value, c.real[i], c.shadow[i])
			}
		}
	}

	t.Run("нет индекса на from", func(t *testing.T) {
		index := series.Series{pt(5, "1000")}
		if _, _, err := ShadowTrades(nil, index, nil, nil, day(1), day(5)); err == nil {
			t.Error("want error")
		}
	})
}
