package service

import (
	"log"
	"time"

	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/CatSprite-dev/fireball/internal/pkg"
)

type PortfolioRequest struct {
	Token      string
	AccountID  string
	OpenedDate time.Time
}

type Calculator struct {
	apiClient *api.Client
}

func NewCalculator(apiClient *api.Client) *Calculator {
	return &Calculator{apiClient: apiClient}
}

func (calc *Calculator) GetFullPortfolio(session PortfolioRequest) (domain.UserFullPortfolio, error) {
	t := time.Now()

	rawPortfolio, err := calc.apiClient.GetPortfolio(session.Token, session.AccountID)
	if err != nil {
		return domain.UserFullPortfolio{}, err
	}
	fullEmptyPortfolio := convertFullPortfolio(rawPortfolio)
	enrichedFullPortfolio, err := enrichFullPortfolio(calc, fullEmptyPortfolio, session.Token, session.AccountID, session.OpenedDate)
	if err != nil {
		return domain.UserFullPortfolio{}, err
	}

	log.Printf("Время выполнения GetFullPortfolio: %.2f сек\n", time.Since(t).Seconds())
	return enrichedFullPortfolio, nil
}

func (calc *Calculator) GetDividends(
	token string,
	accountID string,
	instrumentId string,
	from time.Time,
	to time.Time) (map[string]domain.MoneyValue, error) {

	operations, err := calc.apiClient.GetUserOperationsByCursor(
		token,
		accountID,
		instrumentId,
		&from,
		&to,
		[]pkg.OperationType{pkg.OperationTypeDividend, pkg.OperationTypeCoupon},
		pkg.OperationStateExecuted,
		false,
	)
	if err != nil {
		return nil, err
	}

	result := make(map[string]domain.MoneyValue)
	for _, block := range operations {
		for _, item := range block.Items {
			key := item.Ticker
			if key == "" {
				continue
			}
			current := result[key]
			result[key] = AddMoneyValue(current, domain.MoneyValue(item.Payment))
		}
	}
	return result, nil
}

func (calc *Calculator) GetInstrumentInfo(token string, instrumentIdType pkg.InstrumentIdType, instrumentId string) (domain.Instrument, error) {
	rawInstrument, err := calc.apiClient.GetInstrumentBy(token, instrumentIdType, pkg.ClassCodeUnspecified, instrumentId)
	if err != nil {
		return domain.Instrument{}, err
	}
	instrument := convertInstrument(rawInstrument)
	return instrument, nil
}

func (calc *Calculator) GetIndexByTicker(token string, ticker string) (domain.IndicativeInstruments, error) {
	result := domain.IndicativeInstruments{}

	rawInstruments, err := calc.apiClient.Indicatives(token)
	if err != nil {
		return domain.IndicativeInstruments{}, err
	}
	indicativeInstruments := convertIndicativeInstrument(rawInstruments)
	for _, instr := range indicativeInstruments.Instruments {
		if instr.Ticker == ticker {
			result.Instruments = append(result.Instruments, instr)
		}
	}
	return result, nil
}
