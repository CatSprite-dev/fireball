package metrics

import (
	"testing"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/shopspring/decimal"
)

func TestShadowDeposits(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, time.March, d, 0, 0, 0, 0, time.UTC) }
	pt := func(d int, v string) series.Point {
		return series.Point{Date: day(d), Value: decimal.RequireFromString(v)}
	}

	cases := []struct {
		name     string
		values   series.Series
		index    series.Series
		deposits series.Series
		from, to int
		want     []string // тень на каждый день от from до to
	}{
		{
			name:   "без пополнений",
			values: series.Series{pt(1, "100000")},
			index:  series.Series{pt(1, "1000"), pt(2, "1100")},
			from:   1, to: 2,
			want: []string{"100000", "110000"},
		},
		{
			name:     "пополнение",
			values:   series.Series{pt(1, "100000")},
			index:    series.Series{pt(1, "1000"), pt(2, "1100"), pt(3, "1210")},
			deposits: series.Series{pt(2, "110000")},
			from:     1, to: 3,
			want: []string{"100000", "220000", "242000"},
		},
		{
			name:     "вывод",
			values:   series.Series{pt(1, "100000")},
			index:    series.Series{pt(1, "1000"), pt(2, "1000"), pt(3, "1100")},
			deposits: series.Series{pt(2, "-40000")},
			from:     1, to: 3,
			want: []string{"100000", "60000", "66000"},
		},
		{
			name:   "выходной — индекс прошлого дня",
			values: series.Series{pt(1, "100000")},
			index:  series.Series{pt(1, "1000"), pt(3, "1100")},
			from:   1, to: 3,
			want: []string{"100000", "100000", "110000"},
		},
		{
			name:     "пополнение на from уже в стоимости",
			values:   series.Series{pt(1, "100000")},
			index:    series.Series{pt(1, "1000"), pt(2, "1100")},
			deposits: series.Series{pt(1, "50000")},
			from:     1, to: 2,
			want: []string{"100000", "110000"},
		},
	}
	for _, c := range cases {
		got, err := ShadowDeposits(c.values, c.index, c.deposits, day(c.from), day(c.to))
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if len(got) != len(c.want) {
			t.Errorf("%s: got %d points, want %d: %+v", c.name, len(got), len(c.want), got)
			continue
		}
		for i, w := range c.want {
			if !got[i].Date.Equal(day(c.from+i)) || !got[i].Value.Equal(decimal.RequireFromString(w)) {
				t.Errorf("%s: point %d: %s %s, want %s %s",
					c.name, i, got[i].Date.Format(time.DateOnly), got[i].Value, day(c.from+i).Format(time.DateOnly), w)
			}
		}
	}

	t.Run("нет индекса на from", func(t *testing.T) {
		index := series.Series{pt(5, "1000")}
		if _, err := ShadowDeposits(nil, index, nil, day(1), day(5)); err == nil {
			t.Error("want error")
		}
	})
}
