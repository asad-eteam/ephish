$(document).ready(function () {
    var value = document.cookie;
    console.log('ccccccccccc');
    console.log(document.cookie);
    let name = "_gorilla_csrf";
    getCookie(name)
});

function getCookie(name) {
    var pattern = RegExp(name + "=.[^;]*")
    var matched = document.cookie.match(pattern)
    if (matched) {
        var cookie = matched[0].split('=')
        alert(cookie[1]);
        return cookie[1]
    }
    return false
}