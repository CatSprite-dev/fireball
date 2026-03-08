export interface Quantity {
    units?: string;
    Units?: string;
    nano?: number;
    Nano?: number;
}

export interface Price {
    currency?: string;
    Currency?: string;
    units?: string;
    Units?: string;
    nano?: number;
    Nano?: number;
}

export interface Position {
    positionUid?: string;
    PositionUID?: string;
    name?: string;
    Name?: string;
    ticker?: string;
    Ticker?: string;
    instrumentType?: string;
    InstrumentType?: string;
    quantity?: Quantity;
    averagePositionPrice?: Price;
    currentPrice?: Price;
    dividends?: Price;
    Dividends?: Price;
    totalYield?: Price;
    TotalYield?: Price;
    totalYieldRelative?: Quantity;
    TotalYieldRelative?: Quantity;

}

export interface Portfolio {
    positions: Position[];
    totalAmountPortfolio?: Price;
    expectedYield?: Price;
    totalReturn?: Price;
}

export interface Investment {
    id: string;
    name: string;
    ticker: string;
    quantity: number;
    purchasePrice: number;
    currentPrice: number;
    dividends: number;
    totalYield: number;
    totalYieldRelative: number;
    type: 'stock' | 'etf' | 'crypto' | 'bond' | 'other' | string;
}

export interface InvestmentWithGain extends Investment {
    gain: number;
    gainPercent: number;
}

export interface Metrics {
    totalInvested: number;
    currentValue: number;
    totalGain: number;
    totalGainPercent: number;
    portfolioSize: number;
}

export interface DataPoint {
    invested: number;
    value: number;
}

export interface Candle {
    time: string; // или Date, зависит от того, как приходит
    open: Quantity;
    close: Quantity;
    high?: Quantity; // если добавите позже
    low?: Quantity;  // если добавите позже
    isComplete?: boolean;
}

export interface ChartData {
    index_candles: {
        candles: Candle[];
    };
    // возможно, другие поля
}

declare global {
    interface Window {
        fullPortfolio?: Portfolio;
        deleteInvestment: (id: string) => void;
    }
}