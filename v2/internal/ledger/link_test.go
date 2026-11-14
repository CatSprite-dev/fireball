package ledger

import "testing"

func TestLinkInstruments(t *testing.T) {
	events := []Event{
		{OperationID: "fee-early", ParentID: "buy-2", OperationKind: KindFee},
		{OperationID: "buy-1", OperationKind: KindBuy, InstrumentUID: "share-T"},
		{OperationID: "fee-1", ParentID: "buy-1", OperationKind: KindFee},
		{OperationID: "buy-2", OperationKind: KindBuy, InstrumentUID: "bond-OFZ"},
		{OperationID: "service", OperationKind: KindFee},
		{OperationID: "fee-orphan", ParentID: "missing", OperationKind: KindFee},
		{OperationID: "div", ParentID: "buy-1", OperationKind: KindDividend, InstrumentUID: "share-SBER"},
	}

	InheritInstruments(events)

	want := map[string]string{
		"fee-early":  "bond-OFZ",
		"buy-1":      "share-T",
		"fee-1":      "share-T",
		"buy-2":      "bond-OFZ",
		"service":    "",
		"fee-orphan": "",
		"div":        "share-SBER",
	}
	for _, e := range events {
		if e.InstrumentUID != want[e.OperationID] {
			t.Errorf("%s: got %q, want %q", e.OperationID, e.InstrumentUID, want[e.OperationID])
		}
	}
}
