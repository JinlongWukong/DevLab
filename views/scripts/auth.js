
function loginRequest() {
    var signbtn = document.getElementById("signBtn");
    signbtn.click();
}

function getLoginInfo() {
    var account = window.localStorage.getItem("account");
    var token = window.localStorage.getItem("access_token");
    if (!account || !token) {
        loginRequest();
        throw "login info not found";
    }

    return {
        account,
        token
    };
}

function loadAuth() {
    var loginInfo = getLoginInfo()
    var account = loginInfo.account
    var token = loginInfo.token
    document.getElementById("signBtn").innerText = "Sign out"
    document.getElementById("accountName").innerText = account
    btnAutoClick();
}

function btnAutoClick() {
    console.log("todo by others")
}