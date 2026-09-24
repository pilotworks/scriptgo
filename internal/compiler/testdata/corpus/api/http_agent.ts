// ScriptGo Corpus: Node.js http.Agent Parity Tests
import { Agent, ClientRequest } from "node:http";
import { Socket } from "node:net";

// @api: http.Agent
// @expect: true
const agent = new Agent({
    keepAlive: true,
    maxSockets: 50,
    maxFreeSockets: 10,
    maxTotalSockets: 100,
});
console.log(agent instanceof Agent);

// @api: http.Agent.maxSockets
// @expect: 50
console.log(agent.maxSockets);

// @api: http.Agent.maxFreeSockets
// @expect: 10
console.log(agent.maxFreeSockets);

// @api: http.Agent.maxTotalSockets
// @expect: 100
console.log(agent.maxTotalSockets);

// @api: http.Agent.freeSockets
// @expect: true
console.log(typeof agent.freeSockets === "object" && agent.freeSockets !== null);

// @api: http.Agent.sockets
// @expect: true
console.log(typeof agent.sockets === "object" && agent.sockets !== null);

// @api: http.Agent.requests
// @expect: true
console.log(typeof agent.requests === "object" && agent.requests !== null);

// @api: http.Agent.getName
// @expect: localhost:8080:127.0.0.1
console.log(agent.getName({ host: "localhost", port: 8080, localAddress: "127.0.0.1" }));

// @api: http.Agent.createConnection
// @expect: true
const sock = agent.createConnection({ host: "127.0.0.1", port: 80 });
console.log(sock instanceof Socket);
sock.destroy();

// @api: http.Agent.keepSocketAlive
// @expect: true
const testSock = new Socket();
console.log(agent.keepSocketAlive(testSock));
testSock.destroy();

// @api: http.Agent.reuseSocket
// @expect: true
const req = new ClientRequest({ host: "localhost", path: "/" });
req.on("error", () => {});
const reuseSock = new Socket();
agent.reuseSocket(reuseSock, req);
console.log(req.reusedSocket === true);
reuseSock.destroy();
req.destroy();

// @api: http.Agent.destroy
// @expect: true
agent.destroy();
console.log(Object.keys(agent.sockets).length === 0);
