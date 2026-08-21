console.log("start");

let timerId = 0;
timerId = setInterval(() => {
    console.log("interval fired");
    clearInterval(timerId);
    console.log("cleared interval");
}, 10);
