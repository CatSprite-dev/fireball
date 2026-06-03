package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/CatSprite-dev/fireball/internal/pkg"
	"github.com/CatSprite-dev/fireball/internal/storage"
)

type PortfolioRequest struct {
	ID         string
	Token      string
	AccountID  string
	OpenedDate time.Time
}

var ErrNotFound = errors.New("not found")

var allOperationTypes = []pkg.OperationType{
	pkg.OperationTypeBuy,
	pkg.OperationTypeSell,
	pkg.OperationTypeDividend,
	pkg.OperationTypeCoupon,
	pkg.OperationTypeInput,
	pkg.OperationTypeOutput,
	pkg.OperationTypeInpMulti,
	pkg.OperationTypeOutMulti,
	pkg.OperationTypeBondRepayment,
	pkg.OperationTypeBondRepaymentFull,
}

type Calculator struct {
	APIClient           APIClient
	CandleRepository    *storage.CandleRepository
	OperationRepository *storage.OperationRepository
}

func NewCalculator(apiClient APIClient, candleRepo *storage.CandleRepository, operationsRepo *storage.OperationRepository) *Calculator {
	return &Calculator{
		APIClient:           apiClient,
		CandleRepository:    candleRepo,
		OperationRepository: operationsRepo,
	}
}

func (calc *Calculator) GetFullPortfolio(ctx context.Context, session PortfolioRequest) (domain.Portfolio, error) {
	t := time.Now()

	rawPortfolio, err := calc.APIClient.GetPortfolio(ctx, session.Token, session.AccountID)
	if err != nil {
		return domain.Portfolio{}, err
	}
	portfolio := convertFullPortfolio(rawPortfolio)
	portfolio, err = enrichFullPortfolio(ctx, calc, portfolio, session.Token, session.AccountID, session.OpenedDate)
	if err != nil {
		return domain.Portfolio{}, err
	}

	log.Printf("Время выполнения GetFullPortfolio: %.2f сек\n", time.Since(t).Seconds())
	return portfolio, nil
}

func (calc *Calculator) fetchOperations(
	ctx context.Context,
	token string,
	accountId string,
	instrumentId string,
	from *time.Time,
	to *time.Time,
	operationTypes []pkg.OperationType,
	operationState pkg.OperationState,
	withoutCommissions bool,
) (domain.UserOperations, error) {
	operations, err := calc.APIClient.GetOperationsByCursor(
		ctx, token, accountId, instrumentId, from, to,
		operationTypes, operationState, withoutCommissions,
	)
	if err != nil {
		return domain.UserOperations{}, err
	}
	return convertOperations(operations), nil
}

func (calc *Calculator) fetchAndStoreOperations(
	ctx context.Context,
	token, accountID string,
	instrumentId string,
	from time.Time,
	to time.Time,
	operationTypes []pkg.OperationType,
	operationState pkg.OperationState,
	withoutCommissions bool,
) (domain.UserOperations, error) {
	operations, err := calc.fetchOperations(ctx, token, accountID, instrumentId, &from, &to, allOperationTypes, operationState, withoutCommissions)
	if err != nil {
		return domain.UserOperations{}, err
	}

	opsToReturn := domain.UserOperations{}
	if len(operations.Items) > 0 {
		for _, op := range operations.Items {
			isIn := false
			for _, t := range operationTypes {
				if op.Type == string(t) {
					isIn = true
					break
				}
			}
			if isIn {
				opsToReturn.Items = append(opsToReturn.Items, op)
			}
		}
		err = calc.OperationRepository.PutOperations(ctx, accountID, operations)
		if err != nil {
			log.Printf("error putting fetched operations: %v", err)
		}
	}
	return opsToReturn, nil
}

func (calc *Calculator) GetOrFetchOperations(
	ctx context.Context,
	token string,
	accountId string,
	instrumentId string,
	from time.Time,
	to time.Time,
	operationTypes []pkg.OperationType,
	operationState pkg.OperationState,
	withoutCommissions bool,
) (domain.UserOperations, error) {
	if calc.OperationRepository == nil {
		return calc.fetchOperations(ctx, token, accountId, instrumentId, &from, &to, operationTypes, operationState, withoutCommissions)
	}

	operations, err := calc.OperationRepository.GetOperations(ctx, accountId, from, to, operationTypes)
	if err != nil {
		log.Printf("no operations in db")
		if !errors.Is(err, storage.ErrOperationsNotFound) {
			log.Printf("%v\n", err)
		}
		return calc.fetchAndStoreOperations(ctx, token, accountId, instrumentId, from, to, operationTypes, operationState, withoutCommissions)
	}

	if operations.Items[len(operations.Items)-1].Date.Before(to) {
		from = operations.Items[len(operations.Items)-1].Date.Add(time.Millisecond)
		restOfOperations, err := calc.fetchOperations(ctx, token, accountId, instrumentId, &from, &to, allOperationTypes, operationState, withoutCommissions)
		if err != nil {
			return domain.UserOperations{}, err
		}

		if len(restOfOperations.Items) > 0 {
			for _, op := range restOfOperations.Items {
				isIn := false
				for _, t := range operationTypes {
					if op.Type == string(t) {
						isIn = true
						break
					}
				}
				if isIn {
					operations.Items = append(operations.Items, op)
				}
			}
			err = calc.OperationRepository.PutOperations(ctx, accountId, restOfOperations)
			if err != nil {
				log.Printf("error putting fetched operations: %v", err)
			}
		}
		return operations, nil
	}
	return operations, nil
}

func (calc *Calculator) GetDividends(
	ctx context.Context,
	token string,
	accountID string,
	instrumentId string,
	from time.Time,
	to time.Time,
) (map[string]domain.MoneyValue, error) {
	operations, err := calc.GetOrFetchOperations(
		ctx, token, accountID, instrumentId, from, to,
		[]pkg.OperationType{pkg.OperationTypeDividend, pkg.OperationTypeCoupon},
		pkg.OperationStateExecuted, false,
	)
	if err != nil {
		return nil, err
	}

	result := make(map[string]domain.MoneyValue)
	for _, item := range operations.Items {
		if item.Ticker == "" {
			continue
		}
		result[item.Ticker] = AddMoneyValue(result[item.Ticker], item.Payment)

	}
	return result, nil
}

func (calc *Calculator) GetTotalReturn(
	ctx context.Context,
	token string,
	portfolio domain.Portfolio,
	accountID string,
	openedDate time.Time,
) (totalReturn domain.MoneyValue, totalReturnRelative domain.Quotation, totalInvested domain.MoneyValue, err error) {
	now := time.Now()
	operations, err := calc.GetOrFetchOperations(
		ctx, token, accountID, "", openedDate, now,
		[]pkg.OperationType{
			pkg.OperationTypeInput,
			pkg.OperationTypeOutput,
			pkg.OperationTypeInpMulti,
			pkg.OperationTypeOutMulti,
		},
		pkg.OperationStateExecuted, false,
	)
	if err != nil {
		return domain.MoneyValue{}, domain.Quotation{}, domain.MoneyValue{}, err
	}

	/*candles := []domain.Candle{}
	candles = append(candles, domain.Candle{
		Time: time.Now(),
		Close: domain.Quotation{
			Units: "80", Nano: 0,
		},
		Open: domain.Quotation{
			Units: "80", Nano: 0,
		},
	})*/
	for _, item := range operations.Items {
		if item.Payment.Currency != "rub" {
			totalInvested = calc.ConvertLastPriceToRUB(ctx, token, totalInvested)
		}
		totalInvested = AddMoneyValue(totalInvested, domain.MoneyValue(item.Payment))
	}

	totalReturn = SubtractMoneyValue(portfolio.TotalAmountPortfolio, totalInvested)
	coef, err := DivideMoneyValueToQuotation(totalReturn, totalInvested)
	if err != nil {
		return domain.MoneyValue{}, domain.Quotation{}, domain.MoneyValue{}, err
	}
	totalReturnRelative = MultiplyQuotation(coef, domain.Quotation{Units: "100"})

	return totalReturn, totalReturnRelative, totalInvested, nil
}

func (calc *Calculator) GetInstrument(
	ctx context.Context,
	token string,
	instrumentIdType pkg.InstrumentIdType,
	classCode pkg.ClassCode,
	instrumentId string,
) (domain.Instrument, error) {
	if classCode == "" {
		classCode = pkg.ClassCodeUnspecified
	}
	rawInstrument, err := calc.APIClient.GetInstrumentBy(ctx, token, instrumentIdType, classCode, instrumentId)
	if err != nil {
		var requestErr api.RequestError
		if errors.As(err, &requestErr) && requestErr.StatusCode == http.StatusNotFound {
			return domain.Instrument{}, ErrNotFound
		}
		return domain.Instrument{}, err
	}
	return convertInstrument(rawInstrument), nil
}

func (calc *Calculator) BondBy(
	ctx context.Context,
	token string,
	instrumentIdType pkg.InstrumentIdType,
	classCode pkg.ClassCode,
	instrumentId string,
) (domain.Bond, error) {
	if classCode == "" {
		classCode = pkg.ClassCodeUnspecified
	}

	rawBond, err := calc.APIClient.BondBy(ctx, token, instrumentIdType, classCode, instrumentId)
	if err != nil {
		var requestErr api.RequestError
		if errors.As(err, &requestErr) && requestErr.StatusCode == http.StatusNotFound {
			return domain.Bond{}, ErrNotFound
		}
		return domain.Bond{}, err
	}

	bond := convertBond(rawBond)
	return bond, nil
}

func (calc *Calculator) GetIndexByTicker(ctx context.Context, token string, ticker string) (domain.Instrument, error) {
	rawInstruments, err := calc.APIClient.Indicatives(ctx, token)
	if err != nil {
		return domain.Instrument{}, err
	}
	indicativeInstruments := convertIndicativeInstrument(rawInstruments)
	for _, instr := range indicativeInstruments.Instruments {
		if instr.Ticker == ticker {
			return instr, nil
		}
	}
	return domain.Instrument{}, nil
}

func (calc *Calculator) fetchAndStoreCandles(
	ctx context.Context,
	token, figi string,
	from, to time.Time,
	candleInterval pkg.CandleInterval,
	candleSourceType pkg.CandleSource,
) ([]domain.Candle, error) {
	candles, err := calc.FetchCandles(ctx, token, figi, from, to, candleInterval, candleSourceType)
	if err != nil {
		return nil, err
	}
	if len(candles) > 1 {
		err = calc.CandleRepository.PutCandles(ctx, figi, string(candleInterval), candles[:len(candles)-1])
		if err != nil {
			log.Printf("error putting fetched candles: %v", err)
		}
	}
	return candles, nil
}

func (calc *Calculator) GetOrFetchCandles(
	ctx context.Context,
	token string,
	figi string,
	from time.Time,
	to time.Time,
	candleInterval pkg.CandleInterval,
	candleSourceType pkg.CandleSource,
) ([]domain.Candle, error) {
	if calc.CandleRepository == nil {
		return calc.FetchCandles(ctx, token, figi, from, to, candleInterval, candleSourceType)
	}

	candles, err := calc.CandleRepository.GetCandles(ctx, figi, string(candleInterval), from, to)

	if err != nil {
		if !errors.Is(err, storage.ErrCandlesNotFound) {
			log.Printf("%v\n", err)
		}
		return calc.fetchAndStoreCandles(ctx, token, figi, from, to, candleInterval, candleSourceType)
	}

	if truncateToInterval(candles[len(candles)-1].Time, candleInterval).Before(truncateToInterval(to, candleInterval)) {
		from = truncateToInterval(candles[len(candles)-1].Time, candleInterval).Add(candleIntervalDuration(candleInterval))
		restOfCandles, err := calc.FetchCandles(ctx, token, figi, from, time.Now(), candleInterval, candleSourceType)
		if err != nil {
			return nil, err
		}
		candles = append(candles, restOfCandles...)
		if len(restOfCandles) > 1 {
			err = calc.CandleRepository.PutCandles(ctx, figi, string(candleInterval), restOfCandles[:len(restOfCandles)-1])
			if err != nil {
				log.Printf("error putting fetched candles: %v", err)
			}
		}
		return candles, nil
	}
	return candles, nil
}

func (calc *Calculator) FetchCandles(
	ctx context.Context,
	token string,
	figi string,
	from, to time.Time,
	candleInterval pkg.CandleInterval,
	candleSourceType pkg.CandleSource,
) ([]domain.Candle, error) {
	maxRange := maxIntervalRange(candleInterval)
	var allCandles []domain.Candle

	for chunkFrom := from; chunkFrom.Before(to); {
		chunkTo := chunkFrom.Add(maxRange)
		if chunkTo.After(to) {
			chunkTo = to
		}
		rawCandles, err := calc.APIClient.GetCandles(ctx, token, &chunkFrom, &chunkTo, candleInterval, figi, candleSourceType, 0)
		if err != nil {
			return nil, err
		}
		allCandles = append(allCandles, convertCandles(rawCandles)...)
		chunkFrom = chunkTo
	}

	return allCandles, nil
}

func (calc *Calculator) GetChartData(
	ctx context.Context,
	token string,
	portfolio domain.Portfolio,
	indexTicker string,
	from time.Time,
	to time.Time,
	candleInterval pkg.CandleInterval,
	candleSource pkg.CandleSource,
) (domain.ChartData, error) {
	t := time.Now()
	operations, err := calc.GetOrFetchOperations(
		ctx, token, portfolio.AccountID, "", from, to,
		[]pkg.OperationType{
			pkg.OperationTypeBuy,
			pkg.OperationTypeSell,
			pkg.OperationTypeBondRepaymentFull,
			pkg.OperationTypeBondRepayment,
			pkg.OperationTypeDividend,
			pkg.OperationTypeCoupon,
		},
		pkg.OperationStateExecuted, false,
	)
	if err != nil {
		return domain.ChartData{}, fmt.Errorf("failed to get operations: %w", err)
	}

	opsByInterval := opsByInterval(operations, candleInterval)
	multipliers, nominals := calc.FetchBondInfo(ctx, token, portfolio.Positions, opsByInterval)

	allCurrencies, err := calc.GetCurrencies(ctx, token, pkg.InstrumentStatusAll, pkg.InstrumentExchangeUnspecified)
	if err != nil {
		return domain.ChartData{}, err
	}
	portfolioCurrencies := extractUniqueElements(operations, "currency")
	figis := extractUniqueElements(operations, "figi")

	neededCurrencies := make(map[string]struct{})
	for currency := range portfolioCurrencies {
		if currency != "rub" {
			neededCurrencies[currency] = struct{}{}
		}
	}

	for _, nominal := range nominals {
		if nominal.Currency != "rub" {
			neededCurrencies[nominal.Currency] = struct{}{}
		}
	}

	currencyCandles := make(map[string][]domain.Candle)
	for _, curr := range allCurrencies {
		if curr.ClassCode != string(pkg.ClassCodeEES_CETS) {
			continue
		}
		if _, ok := neededCurrencies[curr.IsoCurrencyName]; !ok {
			continue
		}

		candles, err := calc.GetOrFetchCandles(ctx, token, curr.Figi, from, to, candleInterval, candleSource)
		if err != nil {
			log.Printf(
				"failed to get candles for currency %s: %v", curr.IsoCurrencyName, err)
			continue
		}
		currencyCandles[curr.IsoCurrencyName] = candles
	}

	paymentsByInterval, err := calc.getPaymentsByInterval(operations, currencyCandles, candleInterval)

	if err != nil {
		return domain.ChartData{}, fmt.Errorf("failed to get payments by interval: %w", err)
	}

	index, err := calc.GetIndexByTicker(ctx, token, indexTicker)
	if err != nil {
		return domain.ChartData{}, err
	}

	portfolioCandles, err := calc.BuildPortfolioCandles(ctx, token, portfolio, opsByInterval, paymentsByInterval, currencyCandles, figis, multipliers, nominals, from, to, candleInterval, candleSource)
	if err != nil {
		log.Printf("failed to get portfolio candles: %v", err)
		portfolioCandles = []domain.Candle{}
	}

	indexCandles, err := calc.GetOrFetchCandles(ctx, token, index.Figi, from.AddDate(0, 0, -pkg.CandleFetchBufferDays), to, candleInterval, candleSource)
	if err != nil {
		log.Printf("failed to get index candles: %v", err)
		indexCandles = []domain.Candle{}
	}

	benchmarkCandles, err := buildBenchmarkCandles(opsByInterval, indexCandles, portfolioCandles, paymentsByInterval, candleInterval)
	if err != nil {
		log.Printf("failed to get virtual portfolio candles: %v", err)
		benchmarkCandles = []domain.Candle{}
	}

	times := make([]time.Time, 0, len(portfolioCandles))
	portfolioClose := make([]domain.Quotation, 0, len(portfolioCandles))
	benchmarkClose := make([]domain.Quotation, 0, len(benchmarkCandles))

	for _, c := range portfolioCandles {
		times = append(times, c.Time)
		portfolioClose = append(portfolioClose, c.Close)
	}
	for _, c := range benchmarkCandles {
		benchmarkClose = append(benchmarkClose, c.Close)
	}

	log.Printf("Время выполнения GetChartData: %.2f сек\n", time.Since(t).Seconds())
	return domain.ChartData{
		Times:     times,
		Benchmark: benchmarkClose,
		Portfolio: portfolioClose,
	}, nil
}

// GetCandlesForPortfolio builds historical value candles for the portfolio
// Close = sum(qty * price) for each interval + dividends/coupons received that interval
// Bond prices are converted from % of face value using a multiplier
func (calc *Calculator) BuildPortfolioCandles(
	ctx context.Context,
	token string,
	portfolio domain.Portfolio,
	opsByInterval map[time.Time][]domain.Item,
	paymentsByInterval map[time.Time]domain.MoneyValue,
	currencyCandles map[string][]domain.Candle,
	figis map[string]struct{},
	multipliers map[string]domain.Quotation,
	nominals map[string]domain.MoneyValue,
	from time.Time,
	to time.Time,
	candleInterval pkg.CandleInterval,
	candleSource pkg.CandleSource,
) ([]domain.Candle, error) {
	if portfolio.OpenedDate.After(from) {
		from = portfolio.OpenedDate
	}

	// Reconstruct historical qty per instrument per interval
	historicalHoldings, err := calculateHistoricalHoldings(opsByInterval, portfolio.Positions, nominals, from, to, candleInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical holdings: %w", err)
	}

	historicalCandles, err := calc.FetchHistoricalCandles(ctx, token, figis, from, to, candleInterval, candleSource)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch candles: %w", err)
	}

	// Index candles by truncated interval time for O(1) lookup
	candlesByTime := make(map[string]map[time.Time]domain.Candle)
	for figi, candles := range historicalCandles {
		innerMap := makeCandlesByTime(candles, candleInterval)
		candlesByTime[figi] = innerMap
	}
	intervals := sortedIntervals(historicalHoldings)
	lastPrice := buildLastPriceCache(historicalCandles, from)

	result := make([]domain.Candle, 0, len(intervals))
	for _, interval := range intervals {
		var closeVal domain.Quotation
		for figi, qty := range historicalHoldings[interval] {
			candle, ok := candlesByTime[figi][interval]
			if ok {
				lastPrice[figi] = candle
			}
			if !ok {
				candle, ok = lastPrice[figi]
			}

			close_ := candle.Close
			if m, ok := multipliers[figi]; ok {
				close_ = MultiplyQuotation(close_, m)
				if nominals[figi].Currency != "rub" {
					// Convert to RUB using currency candles
					if currCandles, ok := currencyCandles[nominals[figi].Currency]; ok {
						converted, err := convertToRUB(
							domain.MoneyValue{Units: close_.Units,
								Nano:     close_.Nano,
								Currency: nominals[figi].Currency},
							interval,
							currCandles)
						if err != nil {
							log.Printf("failed to convert %s to RUB for figi %s at interval %v: %v", nominals[figi].Currency, figi, interval, err)
						}
						close_ = domain.Quotation{Units: converted.Units, Nano: converted.Nano}
					}
				}
			}

			posVal := MultiplyQuotation(qty, close_)
			closeVal = AddQuotations(closeVal, posVal)
		}

		if payment, ok := paymentsByInterval[interval]; ok {
			closeVal = AddQuotations(closeVal, domain.Quotation{Units: payment.Units, Nano: payment.Nano})
		}

		result = append(result, domain.Candle{Time: interval, Close: closeVal})
	}

	return result, nil
}

// FetchHistoricalCandlesForPortfolio fetches candles for all figis in parallel.
func (calc *Calculator) FetchHistoricalCandles(
	ctx context.Context,
	token string,
	figis map[string]struct{},
	from time.Time,
	to time.Time,
	candleInterval pkg.CandleInterval,
	candleSource pkg.CandleSource,
) (map[string][]domain.Candle, error) {

	type candleResult struct {
		figi    string
		candles []domain.Candle
		err     error
	}

	resultCh := make(chan candleResult, len(figis))
	for figi := range figis {
		go func(f string) {
			candles, err := calc.GetOrFetchCandles(ctx, token, f, from.AddDate(0, 0, -pkg.CandleFetchBufferDays), to, candleInterval, candleSource)
			resultCh <- candleResult{figi: f, candles: candles, err: err}
		}(figi)
	}

	result := make(map[string][]domain.Candle)
	for range figis {
		res := <-resultCh
		if res.err != nil {
			log.Printf("failed to fetch candles for %s: %v", res.figi, res.err)
			continue
		}
		result[res.figi] = res.candles
	}

	return result, nil
}

// fetchBondMultipliers fetches nominal/100 multiplier for each bond figi in parallel.
// Falls back to 10 (1000 RUB nominal) if bond info is unavailable.
func (calc *Calculator) FetchBondInfo(
	ctx context.Context,
	token string,
	positions []domain.Position,
	operations map[time.Time][]domain.Item,
) (multipliers map[string]domain.Quotation, nominals map[string]domain.MoneyValue) {
	// Collect unique bond figis
	bondFigis := make(map[string]struct{})

	for _, pos := range positions {
		if isBond(pos.InstrumentType) {
			bondFigis[pos.Figi] = struct{}{}
		}
	}

	for _, items := range operations {
		for _, item := range items {
			if isBond(item.InstrumentType) && item.Figi != "" {
				bondFigis[item.Figi] = struct{}{}
			}
		}

	}

	type bondResult struct {
		figi       string
		multiplier domain.Quotation
		nominal    domain.MoneyValue
	}

	resultCh := make(chan bondResult, len(bondFigis))
	for figi := range bondFigis {
		go func(f string) {
			bond, err := calc.BondBy(ctx, token, pkg.InstrumentIdTypeFigi, "", f)
			if err != nil {
				log.Printf("failed to get bond info for %s, using default multiplier: %v", f, err)
				resultCh <- bondResult{figi: f, multiplier: domain.Quotation{Units: "10", Nano: 0}, nominal: domain.MoneyValue{Units: "1000", Nano: 0, Currency: "rub"}}
				return
			}
			multiplier, err := DivideMoneyValue(bond.Nominal, domain.MoneyValue{Units: "100", Nano: 0, Currency: bond.Nominal.Currency})
			if err != nil {
				log.Printf("failed to calculate multiplier for %s, using default: %v", f, err)
				resultCh <- bondResult{figi: f, multiplier: domain.Quotation{Units: "10", Nano: 0}, nominal: domain.MoneyValue{Units: "1000", Nano: 0, Currency: "rub"}}
				return
			}
			resultCh <- bondResult{
				figi:       f,
				multiplier: domain.Quotation{Units: multiplier.Units, Nano: multiplier.Nano},
				nominal:    bond.InitialNominal}
		}(figi)
	}

	multipliers = make(map[string]domain.Quotation, len(bondFigis))
	nominals = make(map[string]domain.MoneyValue, len(bondFigis))
	if len(bondFigis) > 0 {
		for range bondFigis {
			r := <-resultCh
			multipliers[r.figi] = r.multiplier
			nominals[r.figi] = r.nominal
		}
	}
	return multipliers, nominals
}

func (calc *Calculator) GetCurrencies(
	ctx context.Context,
	token string,
	instrumentStatus pkg.InstrumentStatus,
	instrumentExchange pkg.InstrumentExchange) ([]domain.Currency, error) {

	currencies, err := calc.APIClient.Currencies(ctx, token, instrumentStatus, instrumentExchange)
	if err != nil {
		return nil, err
	}

	return converCurrencies(currencies), nil
}

// getPaymentsByInterval sums dividends and coupons received per interval.
func (calc *Calculator) getPaymentsByInterval(
	operations domain.UserOperations,
	currenciesCandles map[string][]domain.Candle,
	candleInterval pkg.CandleInterval,
) (map[time.Time]domain.MoneyValue, error) {

	result := make(map[time.Time]domain.MoneyValue)
	for _, item := range operations.Items {
		switch pkg.OperationType(item.Type) {
		case pkg.OperationTypeDividend, pkg.OperationTypeCoupon:
			interval := truncateToInterval(item.Date, candleInterval)
			if item.Payment.Currency != "rub" {
				candles, ok := currenciesCandles[item.Payment.Currency]
				if !ok {
					log.Printf("no candles found for currency %s, skipping payment conversion", item.Payment.Currency)
					continue
				}
				convertedPayment, err := convertToRUB(item.Payment, item.Date, candles)
				if err != nil {
					log.Printf("failed to convert payment to RUB for currency %s: %v", item.Payment.Currency, err)
					continue
				}
				item.Payment = convertedPayment
			}
			result[interval] = AddMoneyValue(result[interval], item.Payment)
		}
	}
	return result, nil
}

func (calc *Calculator) GetClosePrices(ctx context.Context, token string, instrumentIDs []string, instrumentStatus pkg.InstrumentStatus) ([]domain.ClosePrice, error) {
	raw, err := calc.APIClient.GetClosePrices(ctx, token, instrumentIDs, instrumentStatus)
	if err != nil {
		return []domain.ClosePrice{}, err
	}
	return convertClosePrices(raw), nil
}

func (calc *Calculator) ConvertLastPriceToRUB(ctx context.Context, token string, amount domain.MoneyValue) domain.MoneyValue {
	if amount.Currency == "rub" {
		return amount
	}

	currency := amount.Currency
	currencyID := strings.ToUpper(currency) + strings.ToUpper("rub") + "_" + string(pkg.ClassCodeEES_CETS)
	lastPrices, err := calc.GetClosePrices(ctx, token, []string{currencyID}, pkg.InstrumentStatusAll)
	if err != nil {
		log.Println(err.Error())
	}

	rate := domain.MoneyValue{
		Units: lastPrices[0].ClosePrice.Units,
		Nano:  lastPrices[0].ClosePrice.Nano,
	}

	result := MultiplyMoneyValue(amount, rate)
	result.Currency = "rub"

	return result

}
