const csrfTokenMeta = "csrf-token"

// Run on load
window.onload = function () {
    for (let i = 0; i < 10; i++) {
        loadTweetInID(i)
    }
};

// getMeta get the value of a meta tag
function getMeta(metaName) {
    let metas = document.getElementsByTagName('meta');

    for (let i = 0; i < metas.length; i++) {
        if (metas[i].getAttribute('name') === metaName) {
            return metas[i].getAttribute('content');
        }
    }

    return '';
}

// loadTweetInID load 
function loadTweetInID(id) {

    getTweetID(id)
}

function getTweetID(id) {
    let token = getMeta(csrfTokenMeta)
    var xmlhttp = new XMLHttpRequest();
    var url = "/gettweet?token=" + token;

    xmlhttp.onreadystatechange = function () {
        if (this.readyState == 4 && this.status == 200) {
            var arr = JSON.parse(this.responseText);
            arr['id'] = id
            processTweetID(arr)
        }
    };
    xmlhttp.open("GET", url, false);
    xmlhttp.send();
}

function processTweetID(arr) {
    let data = new Array()

    Object.entries(arr).forEach(([key, value]) => {
        data[key] = value
    });
    // console.log("tweet id " + data['tweetid'])

    if (data['nextloadms'] == "0" || data['nextloadms'] == "") {
        data['nextloadms'] = 30000
    }
    data['id'] = arr['id']


    // console.log("tweet id " + data['tweetid'])

    // return data
    setTweet(data)
}

function setTweet(data) {
    // let tweetElement = document.getElementById("#tweet-" + data['id']);
    console.log("Setting " + data['tweetid'] + " for " + data['id'] + " delay " + data['nextloadms'])

    twttr.widgets
        .createTweet(data['tweetid'], document.getElementById("#tweet-" + data['id']), {
            conversation: "none", // or all
            cards: "hidden", // or visible
            linkColor: "#cc0000", // default is blue
            theme: "light", // or dark
        }).then(function (el) {
            console.log('Tweet added.');
        })

    // Not finished. Need to get and use delay from server
    let delay = data['nextloadms'];

    // Reload same element after delay milliseconds
    // setTimeout(loadTweetInID, delay, data['id']);
    if (data['id'] == 4) {
        setTimeout(window.location.reload.bind(window.location), delay)
    }
}

