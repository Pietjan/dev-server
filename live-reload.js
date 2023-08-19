const evtSource = new EventSource("/dev-server/sse", {
    withCredentials: true,
});

evtSource.onmessage = (event) => {
    window.location.reload(true);
};