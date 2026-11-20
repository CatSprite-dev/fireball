package load

import (
	"slices"
	"testing"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
)

func TestUIDsByKey(t *testing.T) {
	trade := func(instrumentUID, assetUID string, date time.Time) ledger.Event {
		return ledger.Event{
			Date:           date,
			OperationKind:  ledger.KindBuy,
			InstrumentUID:  instrumentUID,
			AssetUID:       assetUID,
			InstrumentKind: ledger.InstrumentShare,
		}
	}
	day := func(d int) time.Time { return time.Date(2026, time.March, d, 12, 0, 0, 0, time.UTC) }

	events := []ledger.Event{
		trade("old-listing", "asset-a", day(1)),
		trade("new-listing", "asset-a", day(5)),
		trade("old-listing", "asset-a", day(3)),
		trade("zeta", "asset-b", day(2)),
		trade("alpha", "asset-b", day(2)),
	}

	got := uidsByKey(events, nil)

	want := map[string][]string{
		"asset-a": {"new-listing", "old-listing"},
		"asset-b": {"alpha", "zeta"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d keys, want %d: %v", len(got), len(want), got)
	}
	for key, w := range want {
		if !slices.Equal(got[key], w) {
			t.Errorf("%s: got %v, want %v", key, got[key], w)
		}
	}
}
