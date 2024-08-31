let debounceTimeout;

function debounce(func, delay) {
    return function (...args) {
        clearTimeout(debounceTimeout);
        debounceTimeout = setTimeout(() => func.apply(this, args), delay);
    };
}

function connect() {
    if (window.liveReloadConnected) {
        console.log("ws-live-reload: already connected");
        return
    }

    window.liveReloadConnected = true;
    var ws = new WebSocket('/__dev-server/ws');

    ws.onmessage = debounce(function (event) {
        console.log(`ws-live-reload: ${event.data}`);
        if (event.data === "reload") {
            window.location.reload(true); // Force reload from server
        }
    }, 500);

    ws.onclose = function (e) {
        window.liveReloadConnected = false;
        setTimeout(function () {
            connect();
        }, 1000);
    };

    ws.onerror = function (err) {
        console.log(`ws-live-reload: ${err}`);
        ws.close();
    };
}

connect();