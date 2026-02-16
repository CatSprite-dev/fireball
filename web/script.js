import { calculatePositionMetrics, calculatePortfolioSummary } from './calculations.js'
import { renderPortfolioHeader, renderPortfolioTable } from './renderers.js'


document.addEventListener('DOMContentLoaded', () => {
  const submitBtn = document.getElementById('submit-btn')
  if (submitBtn) {
    submitBtn.addEventListener('click', checkAndSend)
  }

  const inputField = document.querySelector('.input-field')
  if (inputField) {
    inputField.addEventListener('keypress', (e) => {
      if (e.key === 'Enter') {
        checkAndSend()
      }
    })
  }
})

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
  document.getElementById('start-page').style.display = 'none'
  document.getElementById('main-page').style.display = 'block'
  document.getElementById('portfolio-data').textContent = 'Загрузка...'
  const portfolioData = await authorize(inputText)
  if (portfolioData) {
    showMainPage(portfolioData)
  } else {
    console.log("portfolioData пустой или null")
  }
}

async function checkAndSendWithToken(token) {
  const portfolioData = await authorize(token)
  document.getElementById('start-page').style.display = 'none'
  document.getElementById('main-page').style.display = 'block'
  document.getElementById('portfolio-data').textContent = 'Загрузка...'
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

    console.log("Статус ответа:", response.status)

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
  const fullPortfolio = portfolioData.user_portfolio
  const positions = fullPortfolio.positions
  const allDividends = portfolioData.user_portfolio.allDividends || {}
  
  // Дивиденды по открытым позициям
  const dividends = {}
  positions.forEach(pos => {
    dividends[pos.ticker] = pos.dividends
  })
  
  const positionsMetrics = positions.map(pos => calculatePositionMetrics(pos, dividends))
  
  const summary = calculatePortfolioSummary(positions, dividends, fullPortfolio, allDividends)  // ✅ передаём
  
  const html = `
    ${renderPortfolioHeader(summary)}
    ${renderPortfolioTable(positionsMetrics, summary)}
  `
  
  document.getElementById('portfolio-data').innerHTML = html
}

function invalidToken() {
  console.log("fu and try again")
}