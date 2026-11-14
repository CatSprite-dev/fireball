package ledger

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func withSplit(t *testing.T, asset string, s split) {
	t.Helper()
	splitsByAssetUID[asset] = []split{s}
	t.Cleanup(func() { delete(splitsByAssetUID, asset) })
}

func TestByInstrumentAdjustsForSplit(t *testing.T) {
	const tAsset = "t-asset"
	withSplit(t, tAsset, split{date: time.Date(2026, 4, 17, 0, 0, 0, 0, moscow), before: 1, after: 10})

	before := ev("buy-before", "", KindBuy, "t-uid", "-3107.10", "1")
	before.InstrumentKind = InstrumentShare
	before.AssetUID = tAsset
	before.Date = time.Date(2025, 4, 10, 12, 0, 0, 0, moscow)

	lastDay := ev("buy-last", "", KindBuy, "t-uid", "-3236.20", "1")
	lastDay.InstrumentKind = InstrumentShare
	lastDay.AssetUID = tAsset
	lastDay.Date = time.Date(2026, 4, 14, 18, 0, 0, 0, moscow)

	after := ev("buy-after", "", KindBuy, "t-uid", "-1598", "5")
	after.InstrumentKind = InstrumentShare
	after.AssetUID = tAsset
	after.Date = time.Date(2026, 4, 17, 10, 0, 0, 0, moscow)

	sell := ev("sell-after", "", KindSell, "t-uid", "960", "3")
	sell.InstrumentKind = InstrumentShare
	sell.AssetUID = tAsset
	sell.Date = time.Date(2026, 5, 20, 12, 0, 0, 0, moscow)

	got := ByInstrument([]Event{before, lastDay, after, sell}, nil)

	pos := got[tAsset]
	if !pos.Quantity.Equal(decimal.NewFromInt(22)) {
		t.Errorf("quantity = %s, want 22 (1×10 + 1×10 + 5 − 3)", pos.Quantity)
	}
	if !pos.Bought.Equal(decimal.RequireFromString("-7941.30")) {
		t.Errorf("bought = %s, want -7941.30 (money is not changed by split)", pos.Bought)
	}
}

func TestByInstrumentAdjustsForReverseSplit(t *testing.T) {
	const vtbAsset = "vtbr-asset"
	withSplit(t, vtbAsset, split{date: time.Date(2024, 7, 15, 0, 0, 0, 0, moscow), before: 5000, after: 1})

	before := ev("buy-before", "", KindBuy, "vtbr-uid", "-230", "10000")
	before.InstrumentKind = InstrumentShare
	before.AssetUID = vtbAsset
	before.Date = time.Date(2024, 7, 8, 12, 0, 0, 0, moscow)

	after := ev("buy-after", "", KindBuy, "vtbr-uid", "-1150", "10")
	after.InstrumentKind = InstrumentShare
	after.AssetUID = vtbAsset
	after.Date = time.Date(2024, 7, 15, 12, 0, 0, 0, moscow)

	got := ByInstrument([]Event{before, after}, nil)

	if !got[vtbAsset].Quantity.Equal(decimal.NewFromInt(12)) {
		t.Errorf("quantity = %s, want 12 (10000 / 5000 + 10)", got[vtbAsset].Quantity)
	}
}

func TestAdjustForSplitsExactRatio(t *testing.T) {
	const asset = "test-asset"
	splitDate := time.Date(2025, 1, 10, 0, 0, 0, 0, moscow)
	withSplit(t, asset, split{date: splitDate, before: 3, after: 1})

	got := adjustForSplits(asset, splitDate.AddDate(0, 0, -1), decimal.NewFromInt(300))

	if !got.Equal(decimal.NewFromInt(100)) {
		t.Errorf("quantity = %s, want 100 (300 × 1 / 3)", got)
	}
}
func TestSplitTablesDoNotOverlap(t *testing.T) {
	for uid := range splitsByAssetUID {
		if _, ok := generatedSplitsByAssetUID[uid]; ok {
			t.Errorf("%s is in both tables: split would be applied twice", uid)
		}
	}
}
