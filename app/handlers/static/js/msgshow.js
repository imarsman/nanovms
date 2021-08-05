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
    document.getElementById("#searchtext")
        .addEventListener("keyup", function (event) {
            if (event.keyCode === 13) {
                event.preventDefault();
                // document.getElementById("#searchbutton").click();
                loadSearch()
            }
        });
};

function searchForMessages() {
    loadSearch()
}

function loadSearch() {
    let token = getMeta(csrfTokenMeta)
    var xmlhttp = new XMLHttpRequest();

    let e = document.getElementById("#searchtext");
    let value = e.value
    var url = "/msgsearch?search=" + encodeURIComponent(value)
    // alert("url " + url)

    xmlhttp.onreadystatechange = function () {
        if (this.readyState == 4 && this.status == 200) {
            var resp = this.responseText;
            processResponse(resp)
        }
    };
    xmlhttp.open("GET", url, false);
    xmlhttp.send();
}

function processResponse(resp) {

    // alert("returned")
    console.log(resp)

    let e = document.getElementById("#articles");

    e.innerHTML = resp
}

// function setImageData(data) {
//     let id = data['id']

//     let e = document.getElementById("#image-" + id);

//     let date = data['date']
//     let number = data['number']
//     let title = data['title']
//     let altText = data['alttext']
//     let img = data['img']
//     // Not finished. Need to get and use delay from server
//     let delay = data['nextloadms'];

//     reloadSec = data['nextloadms'] / 1000
//     // &#8212; is an em dash
//     e.innerHTML = "<img title='" + title + "' src='" + img + "' alt='" + altText + "'/>" +
//         "<p style='margin:0px 0px 30px 0px'><small style='font-size: .8em'>" + 
//         date + ' &#8212; "' + title + '"' + " &#8212; reload in " + reloadSec + " sec</small></p>"


//     // console.log("id " + id + " Date " + data['date'] + " for " + data['id'] + " delay " + delay)
//     // Reload same element after delay milliseconds
//     setTimeout(getImageInfo, delay, id)
// }

