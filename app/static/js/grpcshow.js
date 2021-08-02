const csrfTokenMeta = "csrf-token"

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

// Run on load
window.onload = function () {
    for (let i = 0; i < 5; i++) {
        // console.log("loading for " + i)
        loadImgInID(i)
    }
};

// loadTweetInID load 
function loadImgInID(id) {

    // console.log("id " + id)
    getImageInfo(id)
}

function getImageInfo(id) {
    let token = getMeta(csrfTokenMeta)
    var xmlhttp = new XMLHttpRequest();
    var url = "/getimage"

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
        // console.log("key " + key + " value " + value)
    });

    if (data['nextloadms'] == "0" || data['nextloadms'] == "") {
        data['nextloadms'] = 30000
    }
    data['id'] = arr['id']

    setImageData(data)
}

function setImageData(data) {
    let id = data['id']

    let e = document.getElementById("#image-" + id);

    let date = data['date']
    let number = data['number']
    let title = data['title']
    let altText = data['alttext']
    let img = data['img']
    // Not finished. Need to get and use delay from server
    let delay = data['nextloadms'];

    // &#8212; is an em dash
    e.innerHTML = "<img title='" + title + "' src='" + img + "' alt='" + altText + "'/>" +
        "<p style='margin:0px 0px 30px 0px'><small style='font-size: .8em'>" + 
        date + ' &#8212; "' + title + '"' + "</small></p>"


    // console.log("id " + id + " Date " + data['date'] + " for " + data['id'] + " delay " + delay)
    // Reload same element after delay milliseconds
    setTimeout(getImageInfo, delay, id)
}

