// ScriptGo Corpus: Https Standard Builtin APIs
// Consolidated test suite with inline assertions.

import {
    Agent,
    globalAgent,
    ClientRequest,
    Server,
    createServer,
    request,
    get
} from "node:https";
import * as https from "node:https";

// @api: https.Agent
// @expect: 256
// @expect: false
const ag = new Agent();
console.log(ag.maxFreeSockets);
console.log(ag.keepAlive);
ag.destroy();

// @api: https.globalAgent
// @expect: true
console.log(globalAgent instanceof Agent);

// @api: https.ClientRequest
// @expect: https://example.com
// @expect: application/json
// @expect: true
const req = new ClientRequest("https://example.com");
req.setHeader("Content-Type", "application/json");
console.log(req.url);
console.log(req.getHeader("content-type"));
console.log(req.write("data"));
req.setTimeout(5000);
req.end();
req.destroy();

// @api: https.Server
// @api: https.createServer
// @api: server.listen
// @api: server.close
// @api: server.closeAllConnections
// @api: server.closeIdleConnections
// @api: server.setTimeout
// @api: headersTimeout
// @api: keepAliveTimeout
// @api: maxHeadersCount
// @api: requestTimeout
// @api: timeout
// @api: server[Symbol.asyncDispose]()
// @expect: true
// @expect: 60000
// @expect: 5000
// @expect: 2000
// @expect: 300000
// @expect: 60000
// @expect: false
const srv = createServer();
srv.listen(8443);
console.log(srv.listening);
console.log(srv.headersTimeout);
console.log(srv.keepAliveTimeout);
console.log(srv.maxHeadersCount);
console.log(srv.requestTimeout);
srv.closeAllConnections();
srv.closeIdleConnections();
srv.setTimeout(60000);
console.log(srv.timeout);
srv.close();
console.log(srv.listening);
srv.asyncDispose();

// @api: https.request
// @api: https.request(options[, callback])
// @api: https.request(url[, options][, callback])
// @expect: https://api.example.com
const customReq = request("https://api.example.com");
console.log(customReq.url);
customReq.end();

// @api: https.get
// @api: https.get(options[, callback])
// @api: https.get(url[, options][, callback])
// @expect: https://api.github.com
// @expect: true
const getReq = get("https://api.github.com");
console.log(getReq.url);
console.log(getReq.finished);
