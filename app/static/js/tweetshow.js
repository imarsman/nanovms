let items = []

window.onload = function () {

    for (let i = 0; i < 5; i++) {
        loadTweetInID(i)
    }
};

// reload reload a tweet item in delay milliseconds
function loadTweetInID(id) {
    // console.log("loading tweet # " + id)
    var tweet = document.getElementById("#tweet-" + id);

    // Add in getting id here from server
    let tweetID = "1421624040683315200"

    twttr.widgets
        .createTweet(tweetID, tweet, {
            conversation: "none", // or all
            cards: "hidden", // or visible
            linkColor: "#cc0000", // default is blue
            theme: "light", // or dark
        })

    // Not finished. Need to get and use delay from server
    delay = 15000;
    // Reload same element after delay milliseconds
    setTimeout(loadTweetInID, delay, id);
}

