export async function showMainPage(portfolioData) {
  document.getElementById('start-page').style.display = 'none'
  document.getElementById('main-page').style.display = 'block'
  
  const positions = portfolioData.user_portfolio.positions
  const portfolio = portfolioData.user_portfolio
  const total = portfolio.totalAmountPortfolio || {units: "0", currency: "rub"}
  
  const formatCurrency = (val) => {
    const num = parseInt(val?.units || 0)
    return new Intl.NumberFormat('ru-RU').format(num) + ' ₽'
  }
  
  const totalYield = positions.reduce((sum, pos) => 
    sum + parseInt(pos.expectedYield?.units || 0), 0)
  
  let html = `
    <div style="margin-bottom: 20px;">
      <h3 style="margin:0;">Общая стоимость: ${formatCurrency(total)}</h3>  <!-- ИСПРАВЛЕНО -->
      <p style="margin:5px 0 0 0; color:${totalYield >= 0 ? 'green' : 'red'}">
        Общая доходность: ${formatCurrency({units: totalYield})}
      </p>
    </div>
    <table style="width:100%; border-collapse:collapse; font-size:0.9rem;">
      <thead>
        <tr style="background:#f5f5f5;">
          <th style="padding:8px; text-align:left; border-bottom:2px solid #FF2400;">Тикер</th>
          <th style="padding:8px; text-align:right; border-bottom:2px solid #FF2400;">Кол-во</th>
          <th style="padding:8px; text-align:right; border-bottom:2px solid #FF2400;">Средняя</th>
          <th style="padding:8px; text-align:right; border-bottom:2px solid #FF2400;">Текущая</th>
          <th style="padding:8px; text-align:right; border-bottom:2px solid #FF2400; color:${totalYield >= 0 ? 'green' : 'red'}">Доходность</th>
          <th style="padding:8px; text-align:right; border-bottom:2px solid #FF2400;">За день</th>
        </tr>
      </thead>
      <tbody>
  `
  
  positions.forEach((pos, index) => {
    const yieldColor = pos.expectedYield?.units >= 0 ? 'green' : 'red'
    const dailyColor = pos.dailyYield?.units >= 0 ? 'green' : 'red'
    
    html += `
      <tr style="border-bottom:1px solid #eee; ${index % 2 === 0 ? 'background:#fafafa;' : ''}">
        <td style="padding:8px; font-weight:600;">${pos.ticker}</td>
        <td style="padding:8px; text-align:right;">${new Intl.NumberFormat('ru-RU').format(pos.quantity?.units || 0)}</td>
        <td style="padding:8px; text-align:right;">${formatCurrency(pos.averagePositionPrice)}</td>
        <td style="padding:8px; text-align:right; font-weight:500;">${formatCurrency(pos.currentPrice)}</td>
        <td style="padding:8px; text-align:right; color:${yieldColor}">${formatCurrency(pos.expectedYield)}</td>
        <td style="padding:8px; text-align:right; color:${dailyColor}">${formatCurrency(pos.dailyYield)}</td>
      </tr>
    `
  })
  
  html += `
      </tbody>
    </table>
    <div style="margin-top:20px; padding:10px; background:#f8f9fa; border-radius:6px;">
      <p style="margin:0; color:#666;">
        Всего позиций: <strong>${positions.length}</strong> | 
        Обновлено: ${new Date().toLocaleTimeString('ru-RU')}
      </p>
    </div>
  `
  
  document.getElementById('portfolio-data').innerHTML = html
}