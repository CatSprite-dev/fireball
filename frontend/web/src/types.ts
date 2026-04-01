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
    quantity: Quotation
    averagePositionPrice: MoneyValue
    currentPrice: MoneyValue
    expectedYield: MoneyValue
    dividends: MoneyValue
    dailyYield: MoneyValue
    blocked: boolean
    figi: string
    instrumentUid: string
    totalYield: MoneyValue
    totalYieldRelative: Quotation
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
  chart_data: ChartData
}

export interface ChartData {
    times: string[]
    index: Quotation[]
    portfolio: Quotation[]
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
    totalYield: number
    totalYieldRelative: number
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