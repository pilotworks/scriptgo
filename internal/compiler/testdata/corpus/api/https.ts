// ScriptGo Corpus: Https Standard Builtin APIs
// Consolidated test suite with 1:1 isolated assertions for all 17 official Node.js https APIs.

import {
    Agent,
    globalAgent,
    Server,
    createServer,
    request,
    get
} from "node:https";

// @api: new https.Agent
// @expect: agent_inst: true
const ag = new Agent();
console.log("agent_inst: " + (ag.maxFreeSockets === 256));
ag.destroy();

// @api: https.globalAgent
// @expect: globalAgent_inst: true
console.log("globalAgent_inst: " + (globalAgent instanceof Agent));

// @api: new https.Server
// @expect: server_inst: true
const srv1 = new Server();
console.log("server_inst: " + (srv1 instanceof Server));

// @api: https.Server.listen
// @expect: server_listen: true
srv1.listen(8443);
console.log("server_listen: " + srv1.listening);

// @api: https.Server.close
// @expect: server_close: false
srv1.close();
console.log("server_close: " + srv1.listening);

// @api: https.Server.closeAllConnections
// @expect: server_closeAll: true
srv1.closeAllConnections();
console.log("server_closeAll: true");

// @api: https.Server.closeIdleConnections
// @expect: server_closeIdle: true
srv1.closeIdleConnections();
console.log("server_closeIdle: true");

// @api: https.Server.headersTimeout
// @expect: server_headersTimeout: 60000
console.log("server_headersTimeout: " + srv1.headersTimeout);

// @api: https.Server.keepAliveTimeout
// @expect: server_keepAliveTimeout: 5000
console.log("server_keepAliveTimeout: " + srv1.keepAliveTimeout);

// @api: https.Server.maxHeadersCount
// @expect: server_maxHeadersCount: 2000
console.log("server_maxHeadersCount: " + srv1.maxHeadersCount);

// @api: https.Server.requestTimeout
// @expect: server_requestTimeout: 300000
console.log("server_requestTimeout: " + srv1.requestTimeout);

// @api: https.Server.setTimeout
// @expect: server_setTimeout: 60000
srv1.setTimeout(60000);
console.log("server_setTimeout: " + srv1.timeout);

// @api: https.Server.timeout
// @expect: server_timeout: 60000
console.log("server_timeout: " + srv1.timeout);

// @api: https.Server.[Symbol.asyncDispose]
// @expect: server_asyncDispose: true
await srv1[Symbol.asyncDispose]();
console.log("server_asyncDispose: true");

// @api: https.createServer
// @expect: createServer_res: true
const srv2 = createServer();
console.log("createServer_res: " + (srv2 !== null));

// @api: https.request
// @expect: request_res: https://api.example.com
const customReq = request("https://api.example.com");
console.log("request_res: " + customReq.url);
customReq.end();

// @api: https.get
// @expect: get_res: https://api.github.com
// @expect: get_finished: true
const getReq = get("https://api.github.com");
console.log("get_res: " + getReq.url);
console.log("get_finished: " + getReq.finished);

