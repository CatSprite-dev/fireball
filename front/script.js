function checkAndSend() {
  let button = document.querySelector(".submit-btn");
  let input = document.querySelector(".input-field");
  let inputText = input.value;
  if (inputText === "") {
    invalidToken();
  } else {
    authorize();
  }
}

function authorize() {
  console.log("welcome")
}

function invalidToken() {
  console.log("fu and try again")
}