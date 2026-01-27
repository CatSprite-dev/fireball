import { formatCurrency, formatNumberAsCurrency, formatNumberAsPercent } from './formatters.js'

// Расчеты для позиции
export const calculatePositionMetrics = (pos, dividends) => {
  const yieldValue = parseInt(pos.expectedYield?.units || 0) + (parseInt(pos.expectedYield?.nano || 0) / 1_000_000_000)
  const dailyValue = parseInt(pos.dailyYield?.units || 0) + (parseInt(pos.dailyYield?.nano || 0) / 1_000_000_000)
  
  const currentPriceValue = parseFloat(pos.currentPrice?.units || 0) + (parseFloat(pos.currentPrice?.nano || 0) / 1_000_000_000)
  const averagePriceValue = parseFloat(pos.averagePositionPrice?.units || 0) + (parseFloat(pos.averagePositionPrice?.nano || 0) / 1_000_000_000)
  const quantityValue = parseFloat(pos.quantity?.units || 0)
  
  const totalInvestment = averagePriceValue * quantityValue
  const yieldPercent = totalInvestment > 0 ? ((currentPriceValue - averagePriceValue) * quantityValue / totalInvestment) * 100 : 0

  const dividendAmount = dividends[pos.ticker] || 0
  const dividendPercent = totalInvestment > 0 ? (dividendAmount / totalInvestment) * 100 : 0
  
  const totalWithDividends = yieldValue + dividendAmount
  const totalPercent = totalInvestment > 0 ? ((yieldValue + dividendAmount) / totalInvestment) * 100 : 0

  return {
    ticker: pos.ticker,
    quantity: pos.quantity.units,
    averagePrice: formatCurrency(pos.averagePositionPrice),
    currentPrice: formatCurrency(pos.currentPrice),
    yield: {
      value: yieldValue,
      formatted: formatCurrency(pos.expectedYield),
      percent: yieldPercent,
      color: yieldValue >= 0 ? 'green' : 'red'
    },
    daily: {
      value: dailyValue,
      formatted: formatCurrency(pos.dailyYield),
      color: dailyValue >= 0 ? 'green' : 'red'
    },
    dividend: {
      amount: dividendAmount,
      formatted: dividendAmount > 0 ? formatNumberAsCurrency(dividendAmount) : '—',
      percent: dividendPercent,
      color: dividendAmount > 0 ? 'green' : '#666'
    },
    total: {
      value: totalWithDividends,
      formatted: formatNumberAsCurrency(totalWithDividends),
      percent: totalPercent,
      color: totalWithDividends >= 0 ? 'green' : 'red'
    }
  }
}

// Расчеты для всего портфеля
export const calculatePortfolioSummary = (positions, dividends, portfolioTotal) => {
  const totalValue = parseFloat(portfolioTotal?.units || 0) + (parseFloat(portfolioTotal?.nano || 0) / 1_000_000_000)
  
  const totalYield = positions.reduce((sum, pos) => {
    const units = parseInt(pos.expectedYield?.units || 0)
    const nano = parseInt(pos.expectedYield?.nano || 0)
    return sum + units + (nano / 1_000_000_000)
  }, 0)
  
  const totalDividends = Object.values(dividends).reduce((sum, val) => sum + val, 0)
  const totalYieldWithDividends = totalYield + totalDividends
  const totalDailyYield = positions.reduce((sum, pos) => {
    const units = parseInt(pos.dailyYield?.units || 0)
    const nano = parseInt(pos.dailyYield?.nano || 0)
    return sum + units + (nano / 1_000_000_000)
  }, 0)

  return {
    totalValue: formatCurrency(portfolioTotal),
    totalYield: {
      value: totalYield,
      formatted: formatCurrency({
        units: Math.floor(totalYield).toString(),
        nano: Math.round((totalYield - Math.floor(totalYield)) * 1_000_000_000),
        currency: 'rub'
      }),
      percent: totalValue > 0 ? (totalYield / totalValue) * 100 : 0,
      color: totalYield >= 0 ? 'green' : 'red'
    },
    totalDividends: {
      value: totalDividends,
      formatted: formatNumberAsCurrency(totalDividends),
      percent: totalValue > 0 ? (totalDividends / totalValue) * 100 : 0,
      color: totalDividends >= 0 ? 'green' : 'red'
    },
    totalYieldWithDividends: {
      value: totalYieldWithDividends,
      formatted: formatCurrency({
        units: Math.floor(totalYieldWithDividends).toString(),
        nano: Math.round((totalYieldWithDividends - Math.floor(totalYieldWithDividends)) * 1_000_000_000),
        currency: 'rub'
      }),
      percent: totalValue > 0 ? (totalYieldWithDividends / totalValue) * 100 : 0,
      color: totalYieldWithDividends >= 0 ? 'green' : 'red'
    },
    totalDailyYield: {
      value: totalDailyYield,
      formatted: formatNumberAsCurrency(totalDailyYield),
      percent: totalValue > 0 ? (totalDailyYield / totalValue) * 100 : 0,
      color: totalDailyYield >= 0 ? 'green' : 'red'
    },
    positionsCount: positions.length
  }
}