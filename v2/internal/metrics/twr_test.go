package metrics

import (
	"testing"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/shopspring/decimal"
)

func TestTWR(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, time.March, d, 0, 0, 0, 0, time.UTC) }
	pt := func(d int, v string) series.Point {
		return series.Point{Date: day(d), Value: decimal.RequireFromString(v)}
	}

	cases := []struct {
		name     string
		values   series.Series
		in, out  series.Series
		from, to int
		want     []string // рост на каждый день от from до to
	}{
		{
			name:   "пополнение не даёт доходности",
			values: series.Series{pt(1, "100000"), pt(2, "110000"), pt(3, "210000")},
			in:     series.Series{pt(1, "100000"), pt(3, "100000")},
			from:   1, to: 3,
			want: []string{"1", "1.1", "1.1"},
		},
		{
			name:   "первый день: до него пусто, пополнение 1000",
			values: series.Series{pt(2, "1000")},
			in:     series.Series{pt(2, "1000")},
			from:   1, to: 2,
			want: []string{"1", "1"},
		},
		{
			name:   "вывод вечером",
			values: series.Series{pt(1, "100000"), pt(2, "55000")},
			out:    series.Series{pt(2, "-50000")},
			from:   1, to: 2,
			want: []string{"1", "1.05"},
		},
		{
			name:   "пустой портфель без пополнений",
			values: series.Series{pt(1, "0"), pt(3, "0")},
			from:   1, to: 3,
			want: []string{"1", "1", "1"},
		},
		{
			name:   "+10% и −10% подряд",
			values: series.Series{pt(1, "100"), pt(2, "110"), pt(3, "99")},
			from:   1, to: 3,
			want: []string{"1", "1.1", "0.99"},
		},
	}
	for _, c := range cases {
		got := TWR(c.values, c.in, c.out, day(c.from), day(c.to))
		if len(got) != len(c.want) {
			t.Errorf("%s: got %d points, want %d: %+v", c.name, len(got), len(c.want), got)
			continue
		}
		for i, w := range c.want {
			wantDate := day(c.from + i)
			if !got[i].Date.Equal(wantDate) || !got[i].Value.Equal(decimal.RequireFromString(w)) {
				t.Errorf("%s: point %d: %s %s, want %s %s",
					c.name, i, got[i].Date.Format(time.DateOnly), got[i].Value, wantDate.Format(time.DateOnly), w)
			}
		}
	}

	t.Run("приход и уход в один день", func(t *testing.T) {
		values := series.Series{pt(1, "69.51"), pt(2, "0.69")}
		in := series.Series{pt(2, "23000")}
		out := series.Series{pt(2, "-22541")}

		got := TWR(values, in, out, day(1), day(2))

		r := got[1].Value.Sub(decimal.NewFromInt(1))
		if r.Sub(decimal.RequireFromString("-0.0229")).Abs().GreaterThan(decimal.RequireFromString("0.0001")) {
			t.Errorf("r = %s, want ≈ -0.0229", r)
		}
	})
}
