// ScriptGo Corpus: Node.js http.ClientRequest Parity Tests
import { ClientRequest } from "node:http";
import { Socket } from "node:net";

// @api: http.ClientRequest
// @expect: true
const req = new ClientRequest({
    host: "localhost",
    port: 8080,
    path: "/api/test",
    method: "POST"
});
req.on("error", () => {});
console.log(req instanceof ClientRequest);

// @api: http.ClientRequest.host
// @expect: localhost
console.log(req.host);

// @api: http.ClientRequest.path
// @expect: /api/test
console.log(req.path);

// @api: http.ClientRequest.method
// @expect: POST
console.log(req.method);

// @api: http.ClientRequest.protocol
// @expect: http:
console.log(req.protocol);

// @api: http.ClientRequest.maxHeadersCount
// @expect: true
console.log(req.maxHeadersCount === null || typeof req.maxHeadersCount === "number");

// @api: http.ClientRequest.reusedSocket
// @expect: false
console.log(req.reusedSocket);

// @api: http.ClientRequest.aborted
// @expect: false
console.log(req.aborted);

// @api: http.ClientRequest.setHeader
// @expect: application/json
req.setHeader("Content-Type", "application/json");
req.setHeader("X-Custom-Header", "custom-val");
console.log(req.getHeader("content-type"));

// @api: http.ClientRequest.getHeader
// @expect: custom-val
console.log(req.getHeader("x-custom-header"));

// @api: http.ClientRequest.hasHeader
// @expect: true
console.log(req.hasHeader("content-type"));

// @api: http.ClientRequest.getHeaderNames
// @expect: true
const names = req.getHeaderNames();
console.log(names.indexOf("content-type") !== -1 && names.indexOf("x-custom-header") !== -1);

// @api: http.ClientRequest.getRawHeaderNames
// @expect: true
const rawNames = req.getRawHeaderNames();
console.log(rawNames.indexOf("Content-Type") !== -1 && rawNames.indexOf("X-Custom-Header") !== -1);

// @api: http.ClientRequest.getHeaders
// @expect: application/json
const hdrs = req.getHeaders();
console.log(hdrs["content-type"]);

// @api: http.ClientRequest.removeHeader
// @expect: false
req.removeHeader("x-custom-header");
console.log(req.hasHeader("x-custom-header"));

// @api: http.ClientRequest.flushHeaders
// @expect: flush_headers: true
req.flushHeaders();
console.log("flush_headers: true");

// @api: http.ClientRequest.cork
// @expect: cork: true
req.cork();
console.log("cork: true");

// @api: http.ClientRequest.uncork
// @expect: uncork: true
req.uncork();
console.log("uncork: true");

// @api: http.ClientRequest.write
// @expect: true
console.log(req.write("hello-client"));

// @api: http.ClientRequest.setTimeout
// @expect: set_timeout: true
req.setTimeout(5000);
console.log("set_timeout: true");

// @api: http.ClientRequest.setNoDelay
// @expect: set_nodelay: true
req.setNoDelay(true);
console.log("set_nodelay: true");

// @api: http.ClientRequest.setSocketKeepAlive
// @expect: set_keepalive: true
req.setSocketKeepAlive(true, 1000);
console.log("set_keepalive: true");

// @api: http.ClientRequest.writableEnded
// @expect: false
console.log(req.writableEnded);

// @api: http.ClientRequest.writableFinished
// @expect: false
console.log(req.writableFinished);

// @api: http.ClientRequest.finished
// @expect: false
console.log(req.finished);

// @api: http.ClientRequest.socket
// @expect: true
console.log(req.socket === null || req.socket instanceof Socket);

// @api: http.ClientRequest.connection
// @expect: true
console.log(req.connection === req.socket);

// @api: http.ClientRequest.abort
// @expect: true
req.abort();
console.log(req.aborted);

// @api: http.ClientRequest.destroy
// @expect: true
req.destroy();
console.log(typeof req.destroy === "function");

// @api: http.ClientRequest.end
// @expect: true
const req2 = new ClientRequest({ host: "localhost", path: "/" });
req2.on("error", () => {});
req2.end();
console.log(req2.writableEnded);
req2.destroy();
