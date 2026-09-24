// ScriptGo Corpus: Http Standard Builtin APIs
import {
    METHODS,
    STATUS_CODES,
    maxHeaderSize,
    validateHeaderName,
    validateHeaderValue,
    globalAgent,
    createServer,
    get,
    request,
    setMaxIdleHTTPParsers,
    WebSocket,
    Agent
} from "node:http";

// @api: http.METHODS
// @expect: true
// @expect: GET
console.log(METHODS.length > 0);
console.log(METHODS[6]);

// @api: http.STATUS_CODES
// @expect: OK
// @expect: Not Found
console.log(STATUS_CODES["200"]);
console.log(STATUS_CODES["404"]);

// @api: http.validateHeaderName
// @expect: valid
validateHeaderName("Content-Type");
console.log("valid");

// @api: http.validateHeaderValue
// @expect: val-ok
validateHeaderValue("Content-Type", "application/json");
console.log("val-ok");

// @api: http.maxHeaderSize
// @expect: 16384
console.log(maxHeaderSize);

// @api: http.globalAgent
// @expect: true
console.log(globalAgent instanceof Agent);

// @api: http.setMaxIdleHTTPParsers
// @expect: max-parsers-ok
setMaxIdleHTTPParsers(500);
console.log("max-parsers-ok");

// @api: http.createServer
// @expect: true
const srv = createServer();
console.log(typeof srv.listen === "function");
srv.close();

// @api: http.get
// @expect: true
console.log(typeof get === "function");

// @api: http.request
// @expect: true
console.log(typeof request === "function");

// @api: http.WebSocket
// @expect: true
console.log(typeof WebSocket === "function");
