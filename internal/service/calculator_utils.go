package service

import (
	"log"
	"sync"
	"time"

	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/CatSprite-dev/fireball/internal/pkg"
)

func convertToFullPortfolio(raw api.UserPortfolio) domain.UserFullPortfolio {
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

func enrichFullPortfolio(calc *Calculator, fullEmptyPortfolio domain.UserFullPortfolio, token string, accountID string, openedDate time.Time) (domain.UserFullPortfolio, error) {
	// Заполняем ExpectedYield всего портфеля
	coeff, err := DivideMoneyValue(
		domain.MoneyValue{
			Units: fullEmptyPortfolio.ExpectedYieldRelative.Units,
			Nano:  fullEmptyPortfolio.ExpectedYieldRelative.Nano,
		},
		domain.MoneyValue{Units: "100", Nano: 0},
	)
	if err != nil {
		return domain.UserFullPortfolio{}, err
	}
	fullEmptyPortfolio.ExpectedYield = MultiplyMoneyValue(fullEmptyPortfolio.TotalAmountPortfolio, coeff)

	// Получаем дивиденды для всего портфеля
	fullEmptyPortfolio.AllDividends, err = calc.GetDividends(token, accountID, "", openedDate, time.Now().UTC())
	if err != nil {
		return domain.UserFullPortfolio{}, err
	}

	var wg sync.WaitGroup
	for i := range fullEmptyPortfolio.Positions {
		wg.Add(2)
		pos := &fullEmptyPortfolio.Positions[i]

		// Получаем названия инструментов для каждой позиции
		go func(i int, p *domain.Position) {
			defer wg.Done()
			name, err := getInstrumentName(calc.apiClient, token, p.PositionUID)
			if err != nil {
				log.Printf("Failed to get instrument name for %s: %v", p.Ticker, err)
				return
			}
			p.Name = name
		}(i, pos)

		// Заполняем ExpectedYieldRelative, DailyYieldRelative, TotalYield, TotalYieldRelative и дивиденды для каждой позиции
		go func(i int, p *domain.Position) {
			defer wg.Done()
			posAmount := MultiplyMoneyValue(p.AveragePositionPrice,
				domain.MoneyValue{
					Units: p.Quantity.Units,
					Nano:  p.Quantity.Nano,
				},
			)
			p.ExpectedYieldRelative, err = DivideQuotation(
				domain.Quotation{
					Units: p.ExpectedYield.Units,
					Nano:  p.ExpectedYield.Nano,
				},
				domain.Quotation{
					Units: posAmount.Units,
					Nano:  posAmount.Nano,
				},
			)
			if err != nil {
				log.Printf("Failed to calculate ExpectedYieldRelative for %s: %v", p.Ticker, err)
				return
			}
			p.ExpectedYieldRelative = MultiplyQuotation(p.ExpectedYieldRelative, domain.Quotation{Units: "100", Nano: 0})

			// p.DailyYieldRelative =

			divs, err := calc.GetDividends(token, accountID, p.InstrumentUID, openedDate, time.Now().UTC())
			if err != nil {
				log.Printf("Failed to get dividends for %s: %v", p.Ticker, err)
				return
			}
			p.Dividends = divs[p.Ticker]
			p.TotalYield = AddMoneyValue(p.ExpectedYield, p.Dividends)

			p.TotalYieldRelative, err = DivideQuotation(
				domain.Quotation{
					Units: p.TotalYield.Units,
					Nano:  p.TotalYield.Nano,
				},
				domain.Quotation{
					Units: posAmount.Units,
					Nano:  posAmount.Nano,
				},
			)
			if err != nil {
				log.Printf("Failed to calculate TotalYieldRelative for %s: %v", p.Ticker, err)
				return
			}
			p.TotalYieldRelative = MultiplyQuotation(p.TotalYieldRelative, domain.Quotation{Units: "100", Nano: 0})
		}(i, pos)
	}
	wg.Wait()
	return fullEmptyPortfolio, nil
}

func getInstrumentName(apiClient *api.Client, token string, instrumentId string) (string, error) {
	instrument, err := apiClient.GetInstrumentBy(token, pkg.InstrumentIdTypePositionUid, pkg.ClassCodeUnspecified, instrumentId)
	if err != nil {
		return "", err
	}
	return instrument.Instrument.Name, nil
}
