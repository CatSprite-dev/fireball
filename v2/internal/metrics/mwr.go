package metrics

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/series"
)

type cashFlow struct {
	years  float64
	amount float64
}

// MWR — годовая доходность, взвешенная по деньгам (XIRR), за период с from по to.
// Потоки со стороны инвестора: −V(from), −F на каждое пополнение после from, +V(to).
// Пополнения дня from уже внутри V(from) и не учитываются.
// r — результат подбора, а не деньги, поэтому float64; деньги переводятся только внутри расчёта.
func MWR(values, deposits series.Series, from, to time.Time) (float64, error) {
	flows := []cashFlow{{years: 0, amount: -valueAt(values, from).InexactFloat64()}}
	for _, p := range deposits {
		if p.Date.After(from) && !p.Date.After(to) {
			flows = append(flows, cashFlow{years: yearsBetween(from, p.Date), amount: -p.Value.InexactFloat64()})
		}
	}
	flows = append(flows, cashFlow{years: yearsBetween(from, to), amount: valueAt(values, to).InexactFloat64()})

	allZero := true
	for _, f := range flows {
		if f.amount != 0 {
			allZero = false
		}
	}
	if allZero {
		return 0, errors.New("mwr: all cash flows are zero")
	}

	lo, hi := -0.99, 10.0
	fLo, fHi := presentValue(flows, lo), presentValue(flows, hi)
	if fLo*fHi > 0 {
		return 0, fmt.Errorf("mwr: no root in [%g, %g]: present value has the same sign at both ends", lo, hi)
	}
	for i := 0; i < 200 && hi-lo > 1e-9; i++ {
		mid := (lo + hi) / 2
		fMid := presentValue(flows, mid)
		if fMid*fLo > 0 {
			lo, fLo = mid, fMid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2, nil
}

func yearsBetween(from, to time.Time) float64 {
	return to.Sub(from).Hours() / 24 / 365
}

func presentValue(flows []cashFlow, r float64) float64 {
	sum := 0.0
	for _, f := range flows {
		sum += f.amount / math.Pow(1+r, f.years)
	}
	return sum
}
