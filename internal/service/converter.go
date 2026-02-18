package service

import (
	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/domain"
)

func convertFullPortfolio(raw api.UserPortfolio) domain.UserFullPortfolio {
	full := domain.UserFullPortfolio{
		TotalAmountShares:     domain.MoneyValue(raw.TotalAmountShares),
		TotalAmountBonds:      domain.MoneyValue(raw.TotalAmountBonds),
		TotalAmountEtf:        domain.MoneyValue(raw.TotalAmountEtf),
		TotalAmountCurrencies: domain.MoneyValue(raw.TotalAmountCurrencies),
		TotalAmountFutures:    domain.MoneyValue(raw.TotalAmountFutures),
		ExpectedYield:         domain.MoneyValue{},
		ExpectedYieldRelative: domain.Quotation(raw.ExpectedYield),
		AccountID:             raw.AccountID,
		TotalAmountOptions:    domain.MoneyValue(raw.TotalAmountOptions),
		TotalAmountSp:         domain.MoneyValue(raw.TotalAmountSp),
		TotalAmountPortfolio:  domain.MoneyValue(raw.TotalAmountPortfolio),
		DailyYield:            domain.MoneyValue(raw.DailyYield),
		DailyYieldRelative:    domain.Quotation(raw.DailyYieldRelative),
		AllDividends:          map[string]domain.MoneyValue{},
	}

	full.Positions = make([]domain.Position, len(raw.Positions))
	for i, pos := range raw.Positions {
		full.Positions[i] = domain.Position{
			Name:                     "", // из апи мосбиржи
			Type:                     "",
			Figi:                     pos.Figi,
			InstrumentType:           pos.InstrumentType,
			Quantity:                 domain.Quotation(pos.Quantity),
			AveragePositionPrice:     domain.MoneyValue(pos.AveragePositionPrice),
			ExpectedYield:            domain.MoneyValue{Units: pos.ExpectedYield.Units, Nano: pos.ExpectedYield.Nano},
			ExpectedYieldRelative:    domain.Quotation{},
			AveragePositionPricePt:   domain.Quotation(pos.AveragePositionPricePt),
			CurrentPrice:             domain.MoneyValue(pos.CurrentPrice),
			AveragePositionPriceFifo: domain.MoneyValue(pos.AveragePositionPriceFifo),
			QuantityLots:             domain.Quotation(pos.QuantityLots),
			Blocked:                  pos.Blocked,
			BlockedLots:              domain.Quotation(pos.BlockedLots),
			PositionUID:              pos.PositionUID,
			InstrumentUID:            pos.InstrumentUID,
			VarMargin:                domain.MoneyValue(pos.VarMargin),
			ExpectedYieldFifo:        domain.Quotation(pos.ExpectedYieldFifo),
			DailyYield:               domain.MoneyValue(pos.DailyYield),
			DailyYieldRelative:       domain.Quotation{},
			Ticker:                   pos.Ticker,
			ClassCode:                pos.ClassCode,
			CurrentNkd:               domain.MoneyValue(pos.CurrentNkd),
			Dividends:                domain.MoneyValue{},
			TotalYield:               domain.MoneyValue{},
			TotalYieldRelative:       domain.Quotation{},
		}
	}

	return full
}

func convertInstrument(raw api.Instrument) domain.Instrument {
	return domain.Instrument{
		Name:           raw.Instrument.Name,
		InstrumentType: raw.Instrument.InstrumentType,
	}
}

func convertIndicativeInstrument(raw api.IndicativeInstruments) domain.IndicativeInstruments {
	indicatineInstruments := domain.IndicativeInstruments{}
	for _, rawInstr := range raw.Instruments {
		instr := domain.Instrument{
			Figi:           rawInstr.Figi,
			Ticker:         rawInstr.Ticker,
			UID:            rawInstr.UID,
			InstrumentType: rawInstr.InstrumentKind,
			Name:           rawInstr.Name,
		}
		indicatineInstruments.Instruments = append(indicatineInstruments.Instruments, instr)
	}
	return indicatineInstruments
}
