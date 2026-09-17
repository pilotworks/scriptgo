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

// @api: https.get.type
// @expect: true
console.log(typeof get === "function");

// @api: https.request.type
// @expect: true
console.log(typeof request === "function");

// @api: https.globalAgent.maxSockets
// @expect: true
console.log(globalAgent.maxSockets > 0);

// @api: https.request.url
// @expect: true
const req2 = request("https://localhost:8443/status");
req2.on("error", () => {});
console.log(req2.path === "/status");
req2.destroy();


