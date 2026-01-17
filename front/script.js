async function checkAndSend() {
  let button = document.querySelector(".submit-btn");
  let input = document.querySelector(".input-field");
  let inputText = input.value;
  if (inputText === "") {
    invalidToken();
  } else {
    await authorize(inputText);
  }
}

async function authorize(inputText) {
  const response = await fetch("http://localhost:8080/auth", {
    headers: {
      "T-Token": inputText,
    },
  })
  console.log(response.status)
  //set token in local page storage and download new site/refresh current with new assets
  if (response.ok) {
    sessionStorage.setItem('T-Token', inputText)
  }
  showMainPage()
}

async function getPositions() {
  const token = sessionStorage.getItem('T-Token');
  header = ""
  if (!token) {
    header = 0
  }  else {
    header = token
  }
  
  const response = await fetch("http://localhost:8080/portfolio", {
    headers: {
      "T-Token": header,
    },
  })
  if (response.ok) {
    const data = await response.json();
    return data;
  }
  return null;
}

async function showMainPage() {
  document.getElementById('start-page').style.display = 'none'; //  скрывает старую страницу
  document.getElementById('main-page').style.display = 'block'; //  показывает новую
  
  const portfolioData = await getPositions();
  
  if (portfolioData) {
    const formattedData = formatPortfolioData(portfolioData);
    document.getElementById('portfolio-data').innerHTML = formattedData;
  } else {
    document.getElementById('portfolio-data').innerHTML = '<p>Не удалось загрузить данные портфеля</p>';
  }
}

function formatPortfolioData(portfolioData) {
  if (!portfolioData) return '<p>Нет данных</p>';

  const htmlString = portfolioData.replace(/\n/g, '<br>');
  return `<div style="font-family: monospace; white-space: pre-wrap;">${htmlString}</div>`;
}


function invalidToken() {
  console.log("fu and try again")
}
