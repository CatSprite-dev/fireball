package tinvest

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetCurrencies(t *testing.T) {
	var got struct {
		InstrumentStatus string `json:"instrumentStatus"`
	}

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Write([]byte(`{"instruments":[
			{"uid":"a22a1263","ticker":"USD000UTSTOM","classCode":"CETS","isoCurrencyName":"usd",
			 "nominal":{"currency":"usd","units":"1","nano":0}},
			{"uid":"c7c26356","ticker":"KZTRUB_TOM","classCode":"CETS","isoCurrencyName":"kzt",
			 "nominal":{"currency":"kzt","units":"100","nano":0}}
		]}`))
	})

	currencies, err := c.GetCurrencies(context.Background(), InstrumentStatusAll)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.InstrumentStatus != string(InstrumentStatusAll) {
		t.Errorf("request status = %q, want %q", got.InstrumentStatus, InstrumentStatusAll)
	}
	if len(currencies) != 2 {
		t.Fatalf("got %d currencies, want 2", len(currencies))
	}

	kzt := currencies[1]
	if kzt.IsoCurrencyName != "kzt" || kzt.ClassCode != "CETS" || kzt.Nominal.Units != "100" {
		t.Errorf("got %+v, want kzt CETS nominal 100", kzt)
	}
}
