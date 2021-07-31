window.onload = function () {
    var tweet = document.getElementById("tweet");
    var id = tweet.getAttribute("tweetID");

    twttr.widgets
        .createTweet(id, tweet, {
            conversation: "none", // or all
            cards: "hidden", // or visible
            linkColor: "#cc0000", // default is blue
            theme: "light", // or dark
        })
        .then(function (el) {
            el.contentDocument.querySelector(".footer").style.display = "none";
        });
};

// reload reload a tweet item in delay milliseconds
function reload(id) {
    delay = 60000;
    setTimeout(reload, delay, id);
}

