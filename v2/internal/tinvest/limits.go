package tinvest

type group string

const (
	groupAccounts    group = "accounts"
	groupOperations  group = "operations"
	groupInstruments group = "instruments"
	groupMarketData  group = "market_data"
)

const (
	limitAccounts    = 100 // запросов в минуту
	limitOperations  = 200
	limitInstruments = 200
	limitMarketData  = 600
)

var methodGroup = map[string]group{
	"UsersService/GetAccounts":                groupAccounts,
	"OperationsService/GetOperationsByCursor": groupOperations,
	"OperationsService/GetPortfolio":          groupOperations,
	"MarketDataService/GetCandles":            groupMarketData,
}

var groupLimit = map[group]int{
	groupAccounts:    limitAccounts,
	groupOperations:  limitOperations,
	groupInstruments: limitInstruments,
	groupMarketData:  limitMarketData,
}
