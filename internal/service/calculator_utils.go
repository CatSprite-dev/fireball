package service

import (
	"strconv"

	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/pkg"
)

func AddQuotations(a, b pkg.Quotation) pkg.Quotation {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	totalUnits := unitsA + unitsB + int64((a.Nano+b.Nano)/1_000_000_000)
	totalNano := (a.Nano + b.Nano) % 1_000_000_000

	return pkg.Quotation{
		Units: strconv.FormatInt(totalUnits, 10),
		Nano:  totalNano,
	}
}

func convertToFullPortfolio(raw api.UserPortfolio) pkg.UserFullPortfolio {
	full := pkg.UserFullPortfolio{
		TotalAmountShares:     pkg.MoneyValue(raw.TotalAmountShares),
		TotalAmountBonds:      pkg.MoneyValue(raw.TotalAmountBonds),
		TotalAmountEtf:        pkg.MoneyValue(raw.TotalAmountEtf),
		TotalAmountCurrencies: pkg.MoneyValue(raw.TotalAmountCurrencies),
		TotalAmountFutures:    pkg.MoneyValue(raw.TotalAmountFutures),
		ExpectedYield:         pkg.Quotation(raw.ExpectedYield),
		AccountID:             raw.AccountID,
		TotalAmountOptions:    pkg.MoneyValue(raw.TotalAmountOptions),
		TotalAmountSp:         pkg.MoneyValue(raw.TotalAmountSp),
		TotalAmountPortfolio:  pkg.MoneyValue(raw.TotalAmountPortfolio),
		DailyYield:            pkg.MoneyValue(raw.DailyYield),
		DailyYieldRelative:    pkg.Quotation(raw.DailyYieldRelative),
	}

	full.Positions = make([]pkg.Position, len(raw.Positions))
	for i, pos := range raw.Positions {
		full.Positions[i] = pkg.Position{
			Name:                     "", // из апи мосбиржи
			Figi:                     pos.Figi,
			InstrumentType:           pos.InstrumentType,
			Quantity:                 pkg.Quotation(pos.Quantity),
			AveragePositionPrice:     pkg.MoneyValue(pos.AveragePositionPrice),
			ExpectedYield:            pkg.Quotation(pos.ExpectedYield),
			AveragePositionPricePt:   pkg.Quotation(pos.AveragePositionPricePt),
			CurrentPrice:             pkg.MoneyValue(pos.CurrentPrice),
			AveragePositionPriceFifo: pkg.MoneyValue(pos.AveragePositionPriceFifo),
			QuantityLots:             pkg.Quotation(pos.QuantityLots),
			Blocked:                  pos.Blocked,
			BlockedLots:              pkg.Quotation(pos.BlockedLots),
			PositionUID:              pos.PositionUID,
			InstrumentUID:            pos.InstrumentUID,
			VarMargin:                pkg.MoneyValue(pos.VarMargin),
			ExpectedYieldFifo:        pkg.Quotation(pos.ExpectedYieldFifo),
			DailyYield:               pkg.MoneyValue(pos.DailyYield),
			Ticker:                   pos.Ticker,
			ClassCode:                pos.ClassCode,
			CurrentNkd:               pkg.MoneyValue(pos.CurrentNkd),
			Dividends:                pkg.Quotation{},
			TotalYield:               pkg.Quotation{},
		}
	}

	return full
}
