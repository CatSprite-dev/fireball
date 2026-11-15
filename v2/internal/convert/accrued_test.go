package convert

import (
	"testing"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
	"github.com/shopspring/decimal"
)

func TestToAccrued(t *testing.T) {
	list := []tinvest.AccruedInterest{
		{
			Date:    time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC),
			Value:   tinvest.Quotation{Units: "10", Nano: 900000000},
			Nominal: tinvest.Quotation{Units: "1000"},
		},
		{
			Date:    time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC),
			Value:   tinvest.Quotation{Units: "11", Nano: 380000000},
			Nominal: tinvest.Quotation{Units: "1000"},
		},
	}

	nominal, nkd, err := ToAccrued(list)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nominal) != 2 || len(nkd) != 2 {
		t.Fatalf("got %d nominal and %d nkd points, want 2 and 2", len(nominal), len(nkd))
	}

	day := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	if !nkd[1].Date.Equal(day) || !nkd[1].Value.Equal(decimal.RequireFromString("11.38")) {
		t.Errorf("nkd = %+v, want 2026-11-14 11.38", nkd[1])
	}
	if !nominal[1].Date.Equal(day) || !nominal[1].Value.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("nominal = %+v, want 2026-11-14 1000", nominal[1])
	}
}
