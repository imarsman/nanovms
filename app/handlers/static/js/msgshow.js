const csrfTokenMeta = "csrf-token"

// Run on load
window.onload = function () {
    let btn = document.getElementById("#searchtext")
    btn.value = "mites"
    document.getElementById("#searchtext")
        .addEventListener("keyup", function (event) {
            let next = getMeta("search-next")
            if (event.keyCode === 13) {
                event.preventDefault();
                loadSearch(next)
            }
        });
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

function getMetaNode(metaName) {
    let metas = document.getElementsByTagName('meta');

    for (let i = 0; i < metas.length; i++) {
        if (metas[i].getAttribute('name') === metaName) {
            return metas[i]
        }
    }

    return '';
}

function searchForMessages(next) {
    loadSearch(next)
}

function loadSearch(next) {
    let sn = getMetaNode("search-next")
    sn.content = next

    let token = getMeta(csrfTokenMeta)
    var xmlhttp = new XMLHttpRequest();

    let st = document.getElementById("#searchtext");
    let value = st.value
    var url = "/msgsearch?search=" + encodeURIComponent(value) + "&start=" + next

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
