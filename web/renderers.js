import { formatNumberAsPercent } from './formatters.js'

// Рендер заголовка
export const renderPortfolioHeader = (summary) => {
  return `
    <div style="margin-bottom: 20px;">
      <h3 style="margin:0;">Общая стоимость: ${summary.totalValue}</h3>
      <div style="display: flex; gap: 20px; margin-top: 10px; flex-wrap: wrap;">
        <p style="margin:0; color:${summary.totalYieldWithDividends.color}">
          <strong>Общая доходность (с дивидендами):</strong><br>
          ${summary.totalYieldWithDividends.formatted} (${formatNumberAsPercent(summary.totalYieldWithDividends.percent)})
        </p>
        <p style="margin:0; color:${summary.totalYield.color}">
          <strong>От изменения цены:</strong><br>
          ${summary.totalYield.formatted} (${formatNumberAsPercent(summary.totalYield.percent)})
        </p>
        <p style="margin:0; color:${summary.allDividends.color}">
          <strong>Дивиденды (с учетома закрытых позиций):</strong><br>
          ${summary.allDividends.formatted} (${formatNumberAsPercent(summary.allDividends.percent)})
        </p>
      </div>
    </div>
  `
}

// Рендер строки позиции
export const renderPositionRow = (position, index) => {
  return `
    <tr style="border-bottom:1px solid #eee; ${index % 2 === 0 ? 'background:#fafafa;' : ''}">
      <td style="padding:6px; font-weight:600;">${position.name}</td>
      <td style="padding:6px; text-align:right;">${position.quantity}</td>
      <td style="padding:6px; text-align:right;">${position.averagePrice}</td>
      <td style="padding:6px; text-align:right; font-weight:500;">${position.currentPrice}</td>
      <td style="padding:6px; text-align:right; color:${position.yield.color}">
        ${position.yield.formatted} <span style="font-size:0.9em; opacity:0.8;">(${formatNumberAsPercent(position.yield.percent)})</span>
      </td>
      <td style="padding:6px; text-align:right; color:${position.daily.color}">
        ${position.daily.formatted}
      </td>
      <td style="padding:6px; text-align:right; color:${position.dividend.color}">
        ${position.dividend.formatted.includes('—') ? '—' : `${position.dividend.formatted} <span style="font-size:0.9em; opacity:0.8;">(${formatNumberAsPercent(position.dividend.percent)})</span>`}
      </td>
      <td style="padding:6px; text-align:right; color:${position.total.color}; font-weight:500;">
        ${position.total.formatted} <span style="font-size:0.9em; opacity:0.8;">(${formatNumberAsPercent(position.total.percent)})</span>
      </td>
    </tr>
  `
}

// Рендер итоговой строки
export const renderTotalRow = (summary) => {
  return `
    <tr style="background:#f0f8ff; border-top:2px solid #FF2400; font-weight:bold;">
      <td style="padding:8px; text-align:left;">ИТОГО</td>
      <td style="padding:8px; text-align:right;">${summary.positionsCount}</td>
      <td style="padding:8px; text-align:right;">—</td>
      <td style="padding:8px; text-align:right;">—</td>
      <td style="padding:8px; text-align:right; color:${summary.totalYield.color}">
        ${summary.totalYield.formatted} <span style="font-size:0.9em; opacity:0.8;">(${formatNumberAsPercent(summary.totalYield.percent)})</span>
      </td>
      <td style="padding:8px; text-align:right; color:${summary.totalDailyYield.color}">
        ${summary.totalDailyYield.formatted} <span style="font-size:0.9em; opacity:0.8;">(${formatNumberAsPercent(summary.totalDailyYield.percent)})</span>
      </td>
      <td style="padding:8px; text-align:right; color:${summary.openDividends.color}">
        ${summary.openDividends.formatted} <span style="font-size:0.9em; opacity:0.8;">(${formatNumberAsPercent(summary.openDividends.percent)})</span>
      </td>
      <td style="padding:8px; text-align:right; color:${summary.totalOpenYield.color}; font-weight:bold;">
        ${summary.totalOpenYield.formatted} <span style="font-size:0.9em; opacity:0.8;">(${formatNumberAsPercent(summary.totalOpenYield.percent)})</span>
      </td>
    </tr>
  `
}

// Рендер всей таблицы
export const renderPortfolioTable = (positionsMetrics, summary) => {
  let tableHTML = `
    <table style="width:100%; min-width: 900px; border-collapse:collapse; font-size:0.85rem;">
      <thead>
        <tr style="background:#f5f5f5;">
          <th style="padding:6px; text-align:left; border-bottom:2px solid #FF2400;">Тикер</th>
          <th style="padding:6px; text-align:right; border-bottom:2px solid #FF2400;">Кол-во</th>
          <th style="padding:6px; text-align:right; border-bottom:2px solid #FF2400;">Средняя</th>
          <th style="padding:6px; text-align:right; border-bottom:2px solid #FF2400;">Текущая</th>
          <th style="padding:6px; text-align:right; border-bottom:2px solid #FF2400;">Доходность</th>
          <th style="padding:6px; text-align:right; border-bottom:2px solid #FF2400;">За день</th>
          <th style="padding:6px; text-align:right; border-bottom:2px solid #FF2400;">Дивиденды</th>
          <th style="padding:6px; text-align:right; border-bottom:2px solid #FF2400;">Итого</th>
        </tr>
      </thead>
      <tbody>
  `
  
  positionsMetrics.forEach((position, index) => {
    tableHTML += renderPositionRow(position, index)
  })
  
  tableHTML += renderTotalRow(summary)
  
  tableHTML += `
      </tbody>
    </table>
    <div style="margin-top:20px; padding:10px; background:#f8f9fa; border-radius:6px;">
      <p style="margin:0; color:#666;">
        Всего позиций: <strong>${summary.positionsCount}</strong> | 
        Обновлено: ${new Date().toLocaleTimeString('ru-RU')}
      </p>
    </div>
  `
  
  return tableHTML
}