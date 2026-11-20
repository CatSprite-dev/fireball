package metrics

import (
	"math"
	"testing"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/shopspring/decimal"
)

func TestMWR(t *testing.T) {
	from := time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC)
	after := func(days int) time.Time { return from.AddDate(0, 0, days) }
	pt := func(days int, v string) series.Point {
		return series.Point{Date: after(days), Value: decimal.RequireFromString(v)}
	}

	cases := []struct {
		name      string
		values    series.Series
		deposits  series.Series
		days      int
		want      float64
		tolerance float64
	}{
		{
			name:   "год, +10%",
			values: series.Series{pt(0, "100000"), pt(365, "110000")},
			days:   365, want: 0.10, tolerance: 1e-6,
		},
		{
			name:   "30 дней, +1% — в годовых",
			values: series.Series{pt(0, "100000"), pt(30, "101000")},
			days:   30, want: math.Pow(1.01, 365.0/30) - 1, tolerance: 1e-4,
		},
	}
	for _, c := range cases {
		got, err := MWR(c.values, c.deposits, from, after(c.days))
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if math.Abs(got-c.want) > c.tolerance {
			t.Errorf("%s: MWR = %.6f, want %.6f", c.name, got, c.want)
		}
	}

	t.Run("большое пополнение после роста маленькой суммы", func(t *testing.T) {
		values := series.Series{pt(0, "1000"), pt(182, "2000"), pt(365, "102000")}
		deposits := series.Series{pt(182, "100000")}
		got, err := MWR(values, deposits, from, after(365))
		if err != nil {
			t.Fatal(err)
		}
		if got <= 0 || got >= 0.05 {
			t.Errorf("MWR = %.6f, want between 0 and 0.05", got)
		}
	})

	t.Run("вывод", func(t *testing.T) {
		values := series.Series{pt(0, "100000"), pt(365, "55000")}
		deposits := series.Series{pt(182, "-50000")}
		r, err := MWR(values, deposits, from, after(365))
		if err != nil {
			t.Fatal(err)
		}
		sum := -100000 + 50000/math.Pow(1+r, 182.0/365) + 55000/(1+r)
		if math.Abs(sum) > 1e-3 {
			t.Errorf("r = %.6f: сумма потоков %g, want 0", r, sum)
		}
	})

	t.Run("все потоки нулевые", func(t *testing.T) {
		values := series.Series{pt(0, "0"), pt(365, "0")}
		if _, err := MWR(values, nil, from, after(365)); err == nil {
			t.Error("want error")
		}
	})
}
