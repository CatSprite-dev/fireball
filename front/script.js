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
  const response = await fetch("http://localhost:8080/auth/", {
    headers: {
      "T-Token": inputText,
    },
  })
  console.log(response.status)
  //set token in local page storage and download new site/refresh current with new assets
  //sessionStorage.setItem('T-Token', inputText)
  //window.location.replace("second site")
}

function invalidToken() {
  console.log("fu and try again")
}

