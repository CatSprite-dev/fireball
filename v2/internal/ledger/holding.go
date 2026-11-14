package ledger

import "github.com/shopspring/decimal"

type Holding struct {
	AccountID      string
	InstrumentUID  string
	Ticker         string
	InstrumentKind InstrumentKind
	Quantity       decimal.Decimal
	Price          decimal.Decimal
	NKD            decimal.Decimal
	BrokerYield    decimal.Decimal
}

func (h Holding) Value() decimal.Decimal {
	return h.Quantity.Mul(h.Price.Add(h.NKD))
}
