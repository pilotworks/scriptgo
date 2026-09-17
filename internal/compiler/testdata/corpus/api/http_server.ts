// ScriptGo Corpus: Node.js HTTP Server & Client (Strict 1:1 Parity Tests)
import {
    Agent,
    globalAgent,
    Server,
    createServer,
    get,
    request,
    IncomingMessage,
    OutgoingMessage,
    ServerResponse,
    ClientRequest,
    METHODS,
    STATUS_CODES
} from "node:http";

// @api: http.Agent.keepAlive
// @expect: true
const agent = new Agent({ keepAlive: true, maxSockets: 25 });
console.log(agent.keepAlive === true);

// @api: http.Agent.maxSockets
// @expect: true
console.log(agent.maxSockets === 25);

// @api: http.globalAgent
// @expect: true
console.log(globalAgent instanceof Agent);

// @api: http.Server.instanceof
// @expect: true
const server = createServer();
console.log(server instanceof Server);
server.close();

// @api: http.Server.type
// @expect: true
console.log(typeof Server === "function");

// @api: http.Agent.type
// @expect: true
console.log(typeof Agent === "function");

// @api: http.createServer.type
// @expect: true
console.log(typeof createServer === "function");

// @api: http.get.type
// @expect: true
console.log(typeof get === "function");

// @api: http.request.type
// @expect: true
console.log(typeof request === "function");

// @api: http.IncomingMessage.type
// @expect: true
console.log(typeof IncomingMessage === "function");

// @api: http.OutgoingMessage.type
// @expect: true
console.log(typeof OutgoingMessage === "function");

// @api: http.ServerResponse.type
// @expect: true
console.log(typeof ServerResponse === "function");

// @api: http.ClientRequest.type
// @expect: true
console.log(typeof ClientRequest === "function");

// @api: http.METHODS
// @expect: true
console.log(METHODS.indexOf("GET") !== -1 && METHODS.indexOf("POST") !== -1);

// @api: http.STATUS_CODES
// @expect: true
console.log(STATUS_CODES["200"] === "OK" && STATUS_CODES["404"] === "Not Found");

// @api: http.OutgoingMessage.headers
// @expect: true
const out = new OutgoingMessage();
out.setHeader("X-Custom-Header", "hello-world");
console.log(out.getHeader("x-custom-header") === "hello-world");

// @api: http.request.headers
// @expect: true
// @expect: true
const req = request({ host: "localhost", path: "/status", method: "POST" });
req.on("error", () => {});
req.setHeader("X-Client-Req", "active");
console.log(req.hasHeader("x-client-req") === true);
req.removeHeader("x-client-req");
console.log(req.hasHeader("x-client-req") === false);
req.destroy();
