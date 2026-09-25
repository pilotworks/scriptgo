// ScriptGo Corpus: Node.js HTTPS Module (Strict 1:1 Parity Tests)
import {
    Agent,
    globalAgent,
    Server,
    createServer,
    get,
    request
} from "node:https";

// @api: https.Agent.keepAlive
// @expect: true
const agent = new Agent({ keepAlive: true, maxSockets: 50 });
console.log(agent.keepAlive === true);

// @api: https.Agent.maxSockets
// @expect: true
console.log(agent.maxSockets === 50);

// @api: https.globalAgent
// @expect: true
console.log(globalAgent instanceof Agent);

// @api: https.request.headers.get
// @expect: true
const req = request({ host: "localhost", path: "/test", method: "POST" });
req.on("error", () => {});
req.setHeader("X-Custom", "value123");
console.log(req.getHeader("x-custom") === "value123");

// @api: https.request.headers.has
// @expect: true
console.log(req.hasHeader("x-custom") === true);

// @api: https.request.headers.remove
// @expect: true
req.removeHeader("x-custom");
console.log(req.hasHeader("x-custom") === false);
req.destroy();

// @api: https.Server.instanceof
// @expect: true
const server = createServer();
console.log(server instanceof Server);
server.close();

// @api: https.Server.type
// @expect: true
console.log(typeof Server === "function");

// @api: https.Agent.type
// @expect: true
console.log(typeof Agent === "function");

// @api: https.createServer.type
// @expect: true
console.log(typeof createServer === "function");

// @api: https.get
// @expect: true
console.log(typeof get === "function");

// @api: https.request
// @expect: true
const postReq = request({ host: "localhost", path: "/submit", method: "POST" });
postReq.on("error", () => {});
console.log(postReq.method === "POST");
postReq.destroy();

// @api: https.globalAgent.maxSockets
// @expect: true
console.log(globalAgent.maxSockets > 0);

// @api: https.request.url
// @expect: true
const req2 = request("https://localhost:8443/status");
req2.on("error", () => {});
console.log(req2.path === "/status");
req2.destroy();

// @api: https.createServer
// @expect: true
const srv = createServer();
console.log(typeof srv.listen === "function");

// @api: https.https.Agent
// @api: https.Agent
// @expect: true
const testAgent = new Agent({ keepAlive: true });
console.log(testAgent instanceof Agent);

// @api: https.https.Server
// @api: https.Server
// @expect: true
const testSrv = new Server();
console.log(testSrv instanceof Server);

// @api: https.Server.close
// @expect: true
console.log(typeof testSrv.close === "function");

// @api: https.Server.closeAllConnections
// @expect: true
console.log(typeof testSrv.closeAllConnections === "function");

// @api: https.Server.closeIdleConnections
// @expect: true
console.log(typeof testSrv.closeIdleConnections === "function");

// @api: https.Server.listen
// @expect: true
console.log(typeof testSrv.listen === "function");

// @api: https.Server.setTimeout
// @expect: true
testSrv.setTimeout(5000);
console.log(testSrv.timeout === 5000);

// @api: https.Server.headersTimeout
// @expect: true
console.log(testSrv.headersTimeout === 60000);

// @api: https.Server.maxHeadersCount
// @expect: true
testSrv.maxHeadersCount = 2000;
console.log(testSrv.maxHeadersCount === 2000);

// @api: https.Server.requestTimeout
// @expect: true
console.log(testSrv.requestTimeout === 300000);

// @api: https.Server.timeout
// @expect: true
testSrv.timeout = 1000;
console.log(testSrv.timeout === 1000);

// @api: https.Server.keepAliveTimeout
// @expect: true
console.log(testSrv.keepAliveTimeout === 5000);

// @api: https.Server.[Symbol.asyncDispose]
// @expect: true
testSrv.listen(0);
testSrv[Symbol.asyncDispose]();
console.log(true);
