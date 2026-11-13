package convert

import (
	"errors"
	"fmt"

	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
)

func ToHoldings(accountID string, p *tinvest.Portfolio) ([]ledger.Holding, error) {
	result := make([]ledger.Holding, 0, len(p.Positions))

	for _, pos := range p.Positions {
		quantity, err := ToDecimal(pos.Quantity.Units, pos.Quantity.Nano)
		if err != nil {
			return nil, fmt.Errorf("position %s: quantity: %w", pos.Ticker, err)
		}

		price, err := ToDecimal(pos.CurrentPrice.Units, pos.CurrentPrice.Nano)
		if err != nil {
			return nil, fmt.Errorf("position %s: price: %w", pos.Ticker, err)
		}

		nkd, err := ToDecimal(pos.CurrentNkd.Units, pos.CurrentNkd.Nano)
		if err != nil && !errors.Is(err, ErrNoValue) {
			return nil, fmt.Errorf("position %s: nkd: %w", pos.Ticker, err)
		}

		brokerYield, err := ToDecimal(pos.ExpectedYield.Units, pos.ExpectedYield.Nano)
		if err != nil && !errors.Is(err, ErrNoValue) {
			return nil, fmt.Errorf("position %s: expected yield: %w", pos.Ticker, err)
		}

		result = append(result, ledger.Holding{
			AccountID:     accountID,
			InstrumentUID: pos.InstrumentUID,
			Ticker:        pos.Ticker,
			Type:          pos.InstrumentType,
			Quantity:      quantity,
			Price:         price,
			NKD:           nkd,
			BrokerYield:   brokerYield,
		})
	}

	return result, nil
}
