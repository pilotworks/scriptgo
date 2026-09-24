// ScriptGo Corpus: Node.js http.Server Parity Tests
import { Server, createServer } from "node:http";

// @api: http.Server
// @expect: true
const srv = new Server();
console.log(srv instanceof Server);

// @api: http.Server.listen
// @expect: false
// @expect: true
console.log(srv.listening);
srv.listen(0);
console.log(srv.listening);

// @api: http.Server.listening
// @expect: true
console.log(typeof srv.listening === "boolean");

// @api: http.Server.headersTimeout
// @expect: 60000
console.log(srv.headersTimeout);

// @api: http.Server.requestTimeout
// @expect: 300000
console.log(srv.requestTimeout);

// @api: http.Server.maxHeadersCount
// @expect: true
console.log(srv.maxHeadersCount === null || typeof srv.maxHeadersCount === "number");

// @api: http.Server.maxRequestsPerSocket
// @expect: 0
console.log(srv.maxRequestsPerSocket);

// @api: http.Server.timeout
// @expect: 0
console.log(srv.timeout);

// @api: http.Server.keepAliveTimeout
// @expect: 5000
console.log(srv.keepAliveTimeout);

// @api: http.Server.keepAliveTimeoutBuffer
// @expect: true
console.log(srv.keepAliveTimeoutBuffer === undefined || typeof srv.keepAliveTimeoutBuffer === "number");

// @api: http.Server.setTimeout
// @expect: 1000
srv.setTimeout(1000);
console.log(srv.timeout);

// @api: http.Server.closeIdleConnections
// @expect: true
srv.closeIdleConnections();
console.log(typeof srv.closeIdleConnections === "function");

// @api: http.Server.closeAllConnections
// @expect: true
srv.closeAllConnections();
console.log(typeof srv.closeAllConnections === "function");

// @api: http.Server.close
// @expect: true
srv.close();
console.log(typeof srv.close === "function");

// @api: http.Server.[Symbol.asyncDispose]
// @expect: server_asyncDispose: true
const dSrv = new Server();
dSrv.listen(0);
dSrv[Symbol.asyncDispose]();
console.log("server_asyncDispose: true");
