package service

import (
	"time"

	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/domain"
)

func convertFullPortfolio(raw api.UserPortfolio) domain.Portfolio {
	full := domain.Portfolio{
		OpenedDate:            time.Time{},
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
			Name:                     "",
			Type:                     "",
			Figi:                     pos.Figi,
			InstrumentType:           pos.InstrumentType,
			Quantity:                 domain.Quotation(pos.Quantity),
			AveragePositionPrice:     domain.MoneyValue(pos.AveragePositionPrice),
			ExpectedYield:            domain.MoneyValue{Units: pos.ExpectedYield.Units, Nano: pos.ExpectedYield.Nano, Currency: pos.AveragePositionPrice.Currency},
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

func convertInstrument(raw api.InstrumentResponse) domain.Instrument {
	return domain.Instrument{
		Name:           raw.Instrument.Name,
		InstrumentType: raw.Instrument.InstrumentType,
	}
}

func convertBond(raw api.InstrumentBond) domain.Bond {
	return domain.Bond{
		PositionUID:    raw.Bond.PositionUID,
		Name:           raw.Bond.Name,
		Figi:           raw.Bond.Figi,
		UID:            raw.Bond.UID,
		Nominal:        domain.MoneyValue(raw.Bond.Nominal),
		InitialNominal: domain.MoneyValue(raw.Bond.InitialNominal),
		Currency:       raw.Bond.Currency,
		AciValue:       domain.MoneyValue(raw.Bond.AciValue),
		ClassCode:      raw.Bond.ClassCode,
		Ticker:         raw.Bond.Ticker,
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

func convertCandles(raw api.Candles) []domain.Candle {
	candles := []domain.Candle{}
	for _, rawCandle := range raw.Candles {
		candle := domain.Candle{
			Time: rawCandle.Time,
			Close: domain.Quotation{
				Units: rawCandle.Close.Units,
				Nano:  rawCandle.Close.Nano,
			},
			Open: domain.Quotation{
				Units: rawCandle.Open.Units,
				Nano:  rawCandle.Open.Nano,
			},
		}
		candles = append(candles, candle)
	}
	return candles
}

func convertOperations(raw []api.UserOperations) domain.UserOperations {
	result := domain.UserOperations{}
	for _, block := range raw {
		for _, rawItem := range block.Items {
			item := domain.Item{
				BrokerAccountID: rawItem.BrokerAccountID,
				ID:              rawItem.ID,
				InstrumentName:  rawItem.Name,
				Date:            rawItem.Date,
				Type:            rawItem.Type,
				Description:     rawItem.Description,
				State:           rawItem.State,
				InstrumentUID:   rawItem.InstrumentUID,
				Figi:            rawItem.Figi,
				InstrumentType:  rawItem.InstrumentKind,
				PositionUID:     rawItem.PositionUID,
				Ticker:          rawItem.Ticker,
				ClassCode:       rawItem.ClassCode,
				Payment:         domain.MoneyValue(rawItem.Payment),
				InstrumentPrice: domain.MoneyValue(rawItem.Price),
				Commission:      domain.MoneyValue(rawItem.Commission),
				Yield:           domain.MoneyValue(rawItem.Yield),
				YieldRelative:   domain.Quotation(rawItem.YieldRelative),
				AccruedInt:      domain.MoneyValue(rawItem.AccruedInt),
				Quantity:        rawItem.QuantityDone,
			}
			result.Items = append(result.Items, item)
		}
	}
	return result
}

func converCurrencies(raw api.Currencies) []domain.Currency {
	currencies := []domain.Currency{}
	for _, rawCurrency := range raw.Instruments {
		currency := domain.Currency{
			Figi:            rawCurrency.Figi,
			Ticker:          rawCurrency.Ticker,
			ClassCode:       rawCurrency.ClassCode,
			Isin:            rawCurrency.Isin,
			Lot:             rawCurrency.Lot,
			Currency:        rawCurrency.Currency,
			Name:            rawCurrency.Name,
			Nominal:         domain.MoneyValue(rawCurrency.Nominal),
			IsoCurrencyName: rawCurrency.IsoCurrencyName,
			UID:             rawCurrency.UID,
			RealExchange:    rawCurrency.RealExchange,
			PositionUID:     rawCurrency.PositionUID,
			AssetUID:        rawCurrency.AssetUID,
		}
		currencies = append(currencies, currency)
	}
	return currencies
}

func convertClosePrices(raw api.ClosePrices) []domain.ClosePrice {
	closePrices := []domain.ClosePrice{}
	for _, rawClosePrice := range raw.ClosePrices {
		closePrice := domain.ClosePrice{
			Figi:          rawClosePrice.Figi,
			InstrumentUID: rawClosePrice.InstrumentUID,
			Ticker:        rawClosePrice.Ticker,
			ClassCode:     rawClosePrice.ClassCode,
			ClosePrice:    domain.Quotation(rawClosePrice.Price),
			Time:          rawClosePrice.Time,
		}
		closePrices = append(closePrices, closePrice)
	}
	return closePrices
}
