import { formatCurrency, formatNumberAsCurrency, formatNumberAsPercent } from './formatters.js'

export const calculatePositionMetrics = (pos, dividends) => {
  // Дивиденды по тикеру (уже MoneyValue от сервера)
  const dividendAmount = dividends[pos.ticker] || { units: 0, nano: 0, currency: 'rub' }
  
  // Используем готовые значения от сервера
  const yieldValue = parseInt(pos.expectedYield?.units || 0) + (parseInt(pos.expectedYield?.nano || 0) / 1_000_000_000)
  const dailyValue = parseInt(pos.dailyYield?.units || 0) + (parseInt(pos.dailyYield?.nano || 0) / 1_000_000_000)
  const dividendValue = parseInt(dividendAmount.units || 0) + (parseInt(dividendAmount.nano || 0) / 1_000_000_000)
  const totalValue = parseInt(pos.totalYield?.units || 0) + (parseInt(pos.totalYield?.nano || 0) / 1_000_000_000)

  // Проценты уже посчитаны сервером
  const yieldPercent = parseFloat(pos.expectedYieldRelative?.units || 0) + 
                     (parseFloat(pos.expectedYieldRelative?.nano || 0) / 1_000_000_000)

  const totalPercent = parseFloat(pos.totalYieldRelative?.units || 0) + 
                    (parseFloat(pos.totalYieldRelative?.nano || 0) / 1_000_000_000)

  // Стоимость позиции для расчёта процента дивидендов
  const avgPrice = parseFloat(pos.averagePositionPrice?.units || 0) + 
                  (parseFloat(pos.averagePositionPrice?.nano || 0) / 1_000_000_000)
  const quantity = parseFloat(pos.quantity?.units || 0)
  const positionCost = avgPrice * quantity
  
  // Процент дивидендов от стоимости позиции
  const dividendPercent = positionCost > 0 ? (dividendValue / positionCost) * 100 : 0

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
      amount: dividendValue,
      formatted: dividendValue > 0 ? formatNumberAsCurrency(dividendValue) : '—',
      percent: dividendPercent,
      color: dividendValue > 0 ? 'green' : '#666'
    },
    total: {
      value: totalValue,
      formatted: formatCurrency(pos.totalYield),
      percent: totalPercent,
      color: totalValue >= 0 ? 'green' : 'red'
    }
  }
}

export const calculatePortfolioSummary = (positions, dividends, portfolio, allDividends) => {
  const totalValue = parseInt(portfolio.totalAmountPortfolio?.units || 0) + 
                    (parseInt(portfolio.totalAmountPortfolio?.nano || 0) / 1_000_000_000)
  
  // Суммарная доходность от изменения цены (из позиций)
  const totalYieldValue = positions.reduce((sum, pos) => {
    const val = parseInt(pos.expectedYield?.units || 0) + (parseInt(pos.expectedYield?.nano || 0) / 1_000_000_000)
    return sum + val
  }, 0)
  
  // Средний процент доходности от цены
  const totalYieldPercent = totalValue > 0 ? (totalYieldValue / totalValue) * 100 : 0
  
  // Дневная доходность
  const totalDailyYieldValue = positions.reduce((sum, pos) => {
    const val = parseInt(pos.dailyYield?.units || 0) + (parseInt(pos.dailyYield?.nano || 0) / 1_000_000_000)
    return sum + val
  }, 0)
  
  // Дивиденды только по открытым позициям (для итоговой строки)
  const openDividendsValue = positions.reduce((sum, pos) => {
    const val = parseInt(pos.dividends?.units || 0) + (parseInt(pos.dividends?.nano || 0) / 1_000_000_000)
    return sum + val
  }, 0)
  
  // Все дивиденды за всё время (с учетом закрытых позиций)
  const allDividendsValue = Object.values(allDividends || {}).reduce((sum, div) => {
    const val = parseInt(div?.units || 0) + (parseInt(div?.nano || 0) / 1_000_000_000)
    return sum + val
  }, 0)
  
  // Общая доходность с учетом дивидендов от закрытых позиций (для хедера)
  const totalYieldWithDividendsValue = totalYieldValue + allDividendsValue
  const totalYieldWithDividendsPercent = totalValue > 0 ? 
    ((totalYieldValue + allDividendsValue) / totalValue) * 100 : 0

  // Общая доходность только по открытым позициям (для итоговой колонки)
  const totalOpenYieldValue = totalYieldValue + openDividendsValue
  const totalOpenYieldPercent = totalValue > 0 ? 
    ((totalYieldValue + openDividendsValue) / totalValue) * 100 : 0

  return {
    totalValue: formatCurrency(portfolio.totalAmountPortfolio),
    totalYield: {
      value: totalYieldValue,
      formatted: formatNumberAsCurrency(totalYieldValue),
      percent: totalYieldPercent,
      color: totalYieldValue >= 0 ? 'green' : 'red'
    },
    allDividends: {
      value: allDividendsValue,
      formatted: formatNumberAsCurrency(allDividendsValue),
      percent: totalValue > 0 ? (allDividendsValue / totalValue) * 100 : 0,
      color: allDividendsValue >= 0 ? 'green' : 'red'
    },
    openDividends: {
      value: openDividendsValue,
      formatted: formatNumberAsCurrency(openDividendsValue),
      percent: totalValue > 0 ? (openDividendsValue / totalValue) * 100 : 0,
      color: openDividendsValue >= 0 ? 'green' : 'red'
    },
    totalYieldWithDividends: {
      value: totalYieldWithDividendsValue,
      formatted: formatNumberAsCurrency(totalYieldWithDividendsValue),
      percent: totalYieldWithDividendsPercent,
      color: totalYieldWithDividendsValue >= 0 ? 'green' : 'red'
    },
    totalOpenYield: {  // ✅ новая колонка для итоговой строки
      value: totalOpenYieldValue,
      formatted: formatNumberAsCurrency(totalOpenYieldValue),
      percent: totalOpenYieldPercent,
      color: totalOpenYieldValue >= 0 ? 'green' : 'red'
    },
    totalDailyYield: {
      value: totalDailyYieldValue,
      formatted: formatNumberAsCurrency(totalDailyYieldValue),
      percent: totalValue > 0 ? (totalDailyYieldValue / totalValue) * 100 : 0,
      color: totalDailyYieldValue >= 0 ? 'green' : 'red'
    },
    positionsCount: positions.length
  }
}