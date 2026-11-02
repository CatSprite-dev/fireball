package tinvest

type group string

const (
	groupAccounts    group = "accounts"
	groupOperations  group = "operations"
	groupInstruments group = "instruments"
	groupMarketData  group = "market_data"
)

const (
	limitAccounts    = 50 // запросов в минуту, шлюз REST режет до 50, хотя в документации 100
	limitOperations  = 200
	limitInstruments = 200
	limitMarketData  = 600
)

var methodGroup = map[string]group{
	"UsersService/GetAccounts":                groupAccounts,
	"OperationsService/GetOperationsByCursor": groupOperations,
	"OperationsService/GetPortfolio":          groupOperations,
	"MarketDataService/GetCandles":            groupMarketData,
	"InstrumentsService/GetInstrumentBy":      groupInstruments,
	"InstrumentsService/Indicatives":          groupInstruments,
}

var groupLimit = map[group]int{
	groupAccounts:    limitAccounts,
	groupOperations:  limitOperations,
	groupInstruments: limitInstruments,
	groupMarketData:  limitMarketData,
}
