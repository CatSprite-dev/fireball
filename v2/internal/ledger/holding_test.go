package ledger

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestHoldingValue(t *testing.T) {
	h := Holding{
		Quantity: decimal.NewFromInt(200),
		Price:    decimal.RequireFromString("19.8"),
	}
	if got := h.Value(); !got.Equal(decimal.NewFromInt(3960)) {
		t.Errorf("value = %s, want 3960", got)
	}
}
