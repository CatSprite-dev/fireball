window.onload = function() {
  const token = this.sessionStorage.getItem("T-Token")
  if (token) {
    document.getElementById('start-page').style.display = 'none'
    document.getElementById('main-page').style.display = 'block'
    document.getElementById('portfolio-data').textContent = 'Загрузка...'
    checkAndSendWithToken(token)
  }
}

async function checkAndSend() {
  let input = document.querySelector(".input-field")
  let inputText = input.value
  
  if (!inputText) {
    invalidToken()
    return
  }
  const portfolioData = await authorize(inputText)
  if (portfolioData) {
    showMainPage(portfolioData)
  } else {
    console.log("portfolioData пустой или null")
  }
}

async function checkAndSendWithToken(token) {
  const portfolioData = await authorize(token)
  if (portfolioData) {
    showMainPage(portfolioData)
  } else {
    console.log("portfolioData пустой или null в checkAndSendWithToken")
  }
}

async function authorize(inputText) {
  try {
    const response = await fetch("http://localhost:8080/auth", {
      headers: {
        "T-Token": inputText,
      },
    })

    if (response.ok) {
      sessionStorage.setItem('T-Token', inputText)
      const data = await response.json()
      console.log("Данные получены:", data)
      return data
      
    } else {
      console.error(response.status, response.statusText)
      return null
    }
  } catch(error) {
    console.error("Не удалось выполнить запрос", error)
    return null
  }
}

async function showMainPage(portfolioData) {  
  document.getElementById('start-page').style.display = 'none'
  document.getElementById('main-page').style.display = 'block'

  const positions = portfolioData.user_portfolio.positions
  const dividends = portfolioData.user_dividends || {}
  const portfolio = portfolioData.user_portfolio
  const total = portfolio.totalAmountPortfolio || {units: "0", currency: "rub"}
  
  const formatCurrency = (val) => {
    if (!val) return '0 ₽'
    const units = parseInt(val?.units || 0)
    const nano = parseInt(val?.nano || 0)
    const total = units + (nano / 1_000_000_000)
    return new Intl.NumberFormat('ru-RU', {
      minimumFractionDigits: 1,
      maximumFractionDigits: 1
    }).format(total) + ' ₽'
  }

  const formatNumberAsCurrency = (num) => {
    return new Intl.NumberFormat('ru-RU', {
      minimumFractionDigits: 1,
      maximumFractionDigits: 1
    }).format(num) + ' ₽'
  }
  
  // Доходность от изменения цены
  const totalYield = positions.reduce((sum, pos) => {
    const units = parseInt(pos.expectedYield?.units || 0)
    const nano = parseInt(pos.expectedYield?.nano || 0)
    return sum + units + (nano / 1_000_000_000)
  }, 0)
  
  const totalYieldObj = {
    units: Math.floor(totalYield).toString(),
    nano: Math.round((totalYield - Math.floor(totalYield)) * 1_000_000_000),
    currency: 'rub'
  }

  // Доходность от дивидендов
  const totalDividends = Object.values(dividends).reduce((sum, val) => sum + val, 0)
  
  // Общая доходность (цена + дивы)
  const totalYieldWithDividends = totalYield + totalDividends
  const totalYieldWithDividendsObj = {
    units: Math.floor(totalYieldWithDividends).toString(),
    nano: Math.round((totalYieldWithDividends - Math.floor(totalYieldWithDividends)) * 1_000_000_000),
    currency: 'rub'
  }
  
  let html = `
    <div style="margin-bottom: 20px;">
      <h3 style="margin:0;">Общая стоимость: ${formatCurrency(total)}</h3>
      <p style="margin:5px 0 0 0; color:${totalYieldWithDividends >= 0 ? 'green' : 'red'}">
        Общая доходность (цена + дивиденды): ${formatCurrency(totalYieldWithDividendsObj)}
      </p>
      <p style="margin:5px 0 0 0; color:${totalYield >= 0 ? 'green' : 'red'}">
        Доходность от изменения цены: ${formatCurrency(totalYieldObj)}
      </p>
      <p style="margin:5px 0 0 0; color:${totalDividends >= 0 ? 'green' : 'red'}">
        Доходность от дивидендов: ${formatNumberAsCurrency(totalDividends)}
      </p>
    </div>
    <table style="width:100%; min-width: 900px; border-collapse:collapse; font-size:0.85rem;">
      <thead>
        <tr style="background:#f5f5f5;">
          <th style="padding:6px; text-align:left; border-bottom:2px solid #FF2400;">Тикер</th>
          <th style="padding:6px; text-align:right; border-bottom:2px solid #FF2400;">Кол-во</th>
          <th style="padding:6px; text-align:right; border-bottom:2px solid #FF2400;">Средняя</th>
          <th style="padding:6px; text-align:right; border-bottom:2px solid #FF2400;">Текущая</th>
          <th style="padding:6px; text-align:right; border-bottom:2px solid #FF2400; color:${totalYield >= 0 ? 'green' : 'red'}">Доходность</th>
          <th style="padding:6px; text-align:right; border-bottom:2px solid #FF2400;">За день</th>
          <th style="padding:6px; text-align:right; border-bottom:2px solid #FF2400;">Дивиденды</th>
          <th style="padding:6px; text-align:right; border-bottom:2px solid #FF2400;">Итого</th>
        </tr>
      </thead>
      <tbody>
  `
  
  positions.forEach((pos, index) => {    
    const yieldValue = parseInt(pos.expectedYield?.units || 0) + (parseInt(pos.expectedYield?.nano || 0) / 1_000_000_000)
    const dailyValue = parseInt(pos.dailyYield?.units || 0) + (parseInt(pos.dailyYield?.nano || 0) / 1_000_000_000)
    const yieldColor = yieldValue >= 0 ? 'green' : 'red'
    const dailyColor = dailyValue >= 0 ? 'green' : 'red'

    const dividendAmount = dividends[pos.ticker] || 0
    
    const totalWithDividends = yieldValue + dividendAmount
    const dividendColor = dividendAmount > 0 ? 'green' : '#666'
    const totalColor = totalWithDividends >= 0 ? 'green' : 'red'

    html += `
    <tr style="border-bottom:1px solid #eee; ${index % 2 === 0 ? 'background:#fafafa;' : ''}">
      <td style="padding:6px; font-weight:600;">${pos.ticker}</td>
      <td style="padding:6px; text-align:right;">${pos.quantity.units}</td>
      <td style="padding:6px; text-align:right;">${formatCurrency(pos.averagePositionPrice)}</td>
      <td style="padding:6px; text-align:right; font-weight:500;">${formatCurrency(pos.currentPrice)}</td>
      <td style="padding:6px; text-align:right; color:${yieldColor}">${formatCurrency(pos.expectedYield)}</td>
      <td style="padding:6px; text-align:right; color:${dailyColor}">${formatCurrency(pos.dailyYield)}</td>
      <td style="padding:6px; text-align:right; color:${dividendColor}">${dividendAmount > 0 ? formatNumberAsCurrency(dividendAmount) : '—'}</td>
      <td style="padding:6px; text-align:right; color:${totalColor}; font-weight:500;">
        ${formatNumberAsCurrency(totalWithDividends)}
      </td>
    </tr>
  `
  })
  
  // строка итогов
  const totalDailyYield = positions.reduce((sum, pos) => {
    const units = parseInt(pos.dailyYield?.units || 0)
    const nano = parseInt(pos.dailyYield?.nano || 0)
    return sum + units + (nano / 1_000_000_000)
  }, 0)
  
  const totalDailyYieldColor = totalDailyYield >= 0 ? 'green' : 'red'
  const totalDividendsColor = totalDividends >= 0 ? 'green' : 'red'
  const totalOverallColor = totalYieldWithDividends >= 0 ? 'green' : 'red'
  
  html += `
    <tr style="background:#f0f8ff; border-top:2px solid #FF2400; font-weight:bold;">
      <td style="padding:8px; text-align:left;">ИТОГО</td>
      <td style="padding:8px; text-align:right;">${positions.reduce((sum, pos) => sum + parseInt(pos.quantity.units || 0), 0)}</td>
      <td style="padding:8px; text-align:right;">—</td>
      <td style="padding:8px; text-align:right;">—</td>
      <td style="padding:8px; text-align:right; color:${totalYield >= 0 ? 'green' : 'red'}">
        ${formatCurrency(totalYieldObj)}
      </td>
      <td style="padding:8px; text-align:right; color:${totalDailyYieldColor}">
        ${formatNumberAsCurrency(totalDailyYield)}
      </td>
      <td style="padding:8px; text-align:right; color:${totalDividendsColor}">
        ${formatNumberAsCurrency(totalDividends)}
      </td>
      <td style="padding:8px; text-align:right; color:${totalOverallColor}; font-weight:bold;">
        ${formatCurrency(totalYieldWithDividendsObj)}
      </td>
    </tr>
  `
  
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

function invalidToken() {
  console.log("fu and try again")
}