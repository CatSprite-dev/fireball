export interface Quotation{
    units: string
    nano: number
}

export interface MoneyValue {
    currency: string
    units: string
    nano: number
}

export interface Position {
    positionUid: string
    name: string
    ticker: string
    instrumentType: string
    quantity: MoneyValue
    averagePositionPrice: MoneyValue
    currentPrice: MoneyValue
    expectedYield: MoneyValue
    dividends: MoneyValue
}

export interface UserFullPortfolio {
    totalAmountShares: MoneyValue
    totalAmountBonds: MoneyValue
    totalAmountEtf: MoneyValue
    totalAmountCurrencies: MoneyValue
    totalAmountFutures: MoneyValue
    totalAmountOptions: MoneyValue
    totalAmountSp: MoneyValue
    totalAmountPortfolio: MoneyValue
    expectedYield: MoneyValue
    expectedYieldRelative: Quotation
    dailyYield: MoneyValue
    dailyYieldRelative: Quotation
    positions: Position[]
    accountId: string
    allDividends: Record<string, MoneyValue>
}

export interface AuthResponse {
  user_portfolio: UserFullPortfolio
}

export interface Investment {
    id: string
    name: string
    ticker: string
    type: string
    quantity: number
    purchasePrice: number
    currentPrice: number
    dividends: number
}

export interface InvestmentWithGain extends Investment {
    gain: number
    gainPercent: number
}

export interface Metrics {
    totalInvested: number
    currentValue: number
    totalGain: number
    totalGainPercent: number
    dailyYield: number
    dailyYieldRelative: number
    portfolioSize: number
}

export interface DataPoint {
    invested: number
    value: number
}