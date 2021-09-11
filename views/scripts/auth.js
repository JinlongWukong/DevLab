
function getLoginInfo() {
    var account = window.localStorage.getItem("account");
    var token = window.localStorage.getItem("access_token");
    if (!account || !token) {
        throw "login info not found";
    }
    return {
        account,
        token
    };
}
