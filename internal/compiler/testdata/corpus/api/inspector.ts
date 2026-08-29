import {
    Session,
    open,
    close,
    url,
    waitForDebugger,
    inspectorConsole,
    dataReceived,
    dataSent,
    requestWillBeSent,
    responseReceived,
    loadingFinished,
    loadingFailed,
    put
} from "node:inspector";

// @api: inspector.inspector.Session
// @expect: insp_session_inst: true
const sess = new Session();
console.log("insp_session_inst: " + (sess instanceof Session));

// @api: inspector.Session.connect
// @expect: insp_connect: true
sess.connect();
console.log("insp_connect: true");

// @api: inspector.Session.connectToMainThread
// @expect: insp_connectToMain: true
sess.connectToMainThread();
console.log("insp_connectToMain: true");

// @api: inspector.Session.post
// @expect: insp_post: true
sess.post("Runtime.evaluate", {}, (err: unknown, res: unknown) => {});
console.log("insp_post: true");

// @api: inspector.Session.disconnect
// @expect: insp_disconnect: true
sess.disconnect();
console.log("insp_disconnect: true");

// @api: inspector.open
// @expect: insp_open: true
open(9229, "127.0.0.1", false);
console.log("insp_open: true");

// @api: inspector.close
// @expect: insp_close: true
close();
console.log("insp_close: true");

// @api: inspector.url
// @expect: insp_url: undefined
console.log("insp_url: " + url());

// @api: inspector.waitForDebugger
// @expect: insp_wait: true
waitForDebugger();
console.log("insp_wait: true");

// @api: inspector.console
// @expect: insp_console: true
console.log("insp_console: " + (typeof inspectorConsole === "object"));

// @api: inspector.dataReceived
// @expect: insp_dataReceived: true
dataReceived({});
console.log("insp_dataReceived: true");

// @api: inspector.dataSent
// @expect: insp_dataSent: true
dataSent({});
console.log("insp_dataSent: true");

// @api: inspector.requestWillBeSent
// @expect: insp_reqWillSend: true
requestWillBeSent({});
console.log("insp_reqWillSend: true");

// @api: inspector.responseReceived
// @expect: insp_respRecv: true
responseReceived({});
console.log("insp_respRecv: true");

// @api: inspector.loadingFinished
// @expect: insp_loadFin: true
loadingFinished({});
console.log("insp_loadFin: true");

// @api: inspector.loadingFailed
// @expect: insp_loadFail: true
loadingFailed({});
console.log("insp_loadFail: true");

// @api: inspector.put
// @expect: insp_put: true
put({});
console.log("insp_put: true");
