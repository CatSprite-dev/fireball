import { showMainPage } from "./main_page"

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
  }
}

async function checkAndSendWithToken(token) {
  const portfolioData = await authorize(token)
  if (portfolioData) {
    showMainPage(portfolioData)
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



function invalidToken() {
  console.log("fu and try again")
}
