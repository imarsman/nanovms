// reload reload a tweet item in delay milliseconds
function reload(id) {
    delay = 60000;
    setTimeout(reload, delay, id);
}