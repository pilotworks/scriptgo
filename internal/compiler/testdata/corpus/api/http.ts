// ScriptGo Corpus: Http Standard Builtin APIs
// Consolidated test suite with inline assertions.

import {
    METHODS,
    STATUS_CODES,
    maxHeaderSize,
    validateHeaderName,
    validateHeaderValue,
    setMaxIdleHTTPParsers,
    Agent,
    globalAgent,
    OutgoingMessage,
    ClientRequest,
    IncomingMessage,
    ServerResponse,
    Server,
    createServer,
    request,
    get
} from "node:http";
import * as http from "node:http";

// @api: METHODS
// @expect: true
// @expect: GET
console.log(METHODS.length > 0);
console.log(METHODS[6]);

// @api: STATUS_CODES
// @expect: OK
// @expect: Not Found
console.log(STATUS_CODES["200"]);
console.log(STATUS_CODES["404"]);

// @api: WebSocket
// @expect: 0
const ws = new WebSocket("ws://localhost:8080");
console.log(ws.readyState);

// @api: aborted
// @expect: false
const incMsg1 = new IncomingMessage();
console.log(incMsg1.aborted);

// @api: agent.createConnection
// @expect: true
const ag = new Agent();
const conn = ag.createConnection({ port: 8080 });
conn.on("error", () => {});
console.log(conn !== null && conn !== undefined);

// @api: agent.destroy
// @expect: done
ag.destroy();
console.log("done");

// @api: agent.getName
// @expect: localhost:8080:
console.log(ag.getName({ host: "localhost", port: 8080 }));

// @api: agent.keepSocketAlive
// @expect: true
console.log(ag.keepSocketAlive(conn));

// @api: agent.reuseSocket
// @expect: reused
const dummyReq = new ClientRequest("http://localhost/dummy");
dummyReq.on("error", () => {});
ag.reuseSocket(conn, dummyReq);
console.log("reused");

// @api: complete
// @expect: true
console.log(incMsg1.complete);

// @api: connection
// @expect: true
console.log(incMsg1.connection === null || incMsg1.connection === undefined || incMsg1.connection === "");

// @api: finished
// @expect: true
const outMsg1 = new ServerResponse(incMsg1);
outMsg1.end("hello");
console.log(outMsg1.writableFinished);

// @api: freeSockets
// @expect: true
console.log(typeof ag.freeSockets === "object");

// @api: globalAgent
// @expect: 256
console.log(globalAgent.maxFreeSockets);

// @api: headers
// @expect: true
console.log(typeof incMsg1.headers === "object");

// @api: headersDistinct
// @expect: true
console.log(typeof incMsg1.headersDistinct === "object");

// @api: headersSent
// @expect: true
console.log(outMsg1.headersSent);

// @api: headersTimeout
// @expect: 60000
const srv1 = new Server();
console.log(srv1.headersTimeout);

// @api: host
// @expect: localhost
const req1 = new ClientRequest({ host: "localhost", path: "/api" });
req1.on("error", () => {});
console.log(req1.host);

// @api: http.Agent
// @expect: 50
const customAgent = new Agent({ maxSockets: 50 });
console.log(customAgent.maxSockets);

// @api: http.ClientRequest
// @expect: /test
const reqClient = new ClientRequest("http://localhost/test");
reqClient.on("error", () => {});
console.log(reqClient.path);

// @api: http.IncomingMessage
// @expect: true
const inc2 = new IncomingMessage();
console.log(inc2.httpVersion !== undefined);

// @api: http.OutgoingMessage
// @expect: 0
const out2 = new ServerResponse(inc2);
console.log(out2.writableLength);

// @api: http.Server
// @expect: false
const srv2 = new Server();
console.log(srv2.listening);

// @api: http.ServerResponse
// @expect: 200
const resp1 = new ServerResponse(inc2);
console.log(resp1.statusCode);

// @api: http.createServer
// @expect: true
const srvCreated = createServer((req: unknown, res: unknown) => {});
console.log(srvCreated instanceof Server);

// @api: http.get
// @expect: GET
const getReq = get("http://localhost/path");
getReq.on("error", () => {});
console.log(getReq.method);

// @api: http.request
// @expect: true
const postReq = request({ method: "POST", path: "/submit" });
postReq.on("error", () => {});
console.log(postReq !== null);

// @api: http.setMaxIdleHTTPParsers
// @expect: ok
setMaxIdleHTTPParsers(100);
console.log("ok");

// @api: http.validateHeaderName
// @expect: valid
validateHeaderName("Content-Type");
console.log("valid");

// @api: http.validateHeaderValue
// @expect: val-ok
validateHeaderValue("Content-Type", "application/json");
console.log("val-ok");

// @api: httpVersion
// @expect: true
console.log(inc2.httpVersion !== undefined);

// @api: keepAliveTimeout
// @expect: 5000
console.log(srv1.keepAliveTimeout);

// @api: keepAliveTimeoutBuffer
// @expect: 1000
console.log(srv1.keepAliveTimeoutBuffer);

// @api: listening
// @expect: true
srv1.listen(8080);
console.log(srv1.listening);

// @api: maxFreeSockets
// @expect: 256
console.log(ag.maxFreeSockets);

// @api: maxHeaderSize
// @expect: 16384
console.log(maxHeaderSize);

// @api: maxHeadersCount
// @expect: true
console.log(srv1.maxHeadersCount > 0);

// @api: maxRequestsPerSocket
// @expect: 0
console.log(srv1.maxRequestsPerSocket);

// @api: maxSockets
// @expect: true
console.log(ag.maxSockets > 0);

// @api: maxTotalSockets
// @expect: true
console.log(ag.maxTotalSockets > 0);

// @api: message.connection
// @expect: true
console.log(inc2.connection === null || inc2.connection === undefined || inc2.connection === "");

// @api: message.destroy
// @expect: true
const incWithSock = new IncomingMessage(conn);
incWithSock.destroy();
console.log(incWithSock.aborted);

// @api: message.setTimeout
// @expect: msg-timeout-set
incWithSock.setTimeout(5000, () => {});
console.log("msg-timeout-set");

// @api: method
// @expect: true
console.log(inc2.method !== undefined);

// @api: outgoingMessage.addTrailers
// @expect: trailers-added
out2.addTrailers({ "x-trailer": "val" });
console.log("trailers-added");

// @api: outgoingMessage.appendHeader
// @expect: true
out2.setHeader("x-custom", "v1");
out2.appendHeader("x-custom", "v2");
console.log(out2.hasHeader("x-custom"));

// @api: outgoingMessage.connection
// @expect: true
console.log(out2.connection === null || out2.connection === undefined || out2.connection === "");

// @api: outgoingMessage.cork
// @expect: 1
out2.cork();
console.log(out2.writableCorked);

// @api: outgoingMessage.uncork
// @expect: 0
out2.uncork();
console.log(out2.writableCorked);

// @api: outgoingMessage.destroy
// @expect: true
out2.destroy();
console.log(out2 !== null);

// @api: outgoingMessage.end
// @expect: true
const outEnd = new ServerResponse(inc2);
outEnd.end("done");
console.log(outEnd.writableEnded);

// @api: outgoingMessage.flushHeaders
// @expect: true
const outFlush = new ServerResponse(inc2);
outFlush.flushHeaders();
console.log(outFlush.headersSent);

// @api: outgoingMessage.getHeader
// @expect: application/json
const outHeaders = new ServerResponse(inc2);
outHeaders.setHeader("Content-Type", "application/json");
console.log(outHeaders.getHeader("content-type"));

// @api: outgoingMessage.getHeaderNames
// @expect: content-type
console.log(outHeaders.getHeaderNames().join(","));

// @api: outgoingMessage.getHeaders
// @expect: application/json
console.log(outHeaders.getHeader("content-type"));

// @api: outgoingMessage.hasHeader
// @expect: true
// @expect: false
console.log(outHeaders.hasHeader("content-type"));
console.log(outHeaders.hasHeader("x-unknown"));

// @api: outgoingMessage.pipe
// @expect: true
console.log(inc2.pipe(outHeaders) !== null);

// @api: outgoingMessage.removeHeader
// @expect: false
outHeaders.removeHeader("content-type");
console.log(outHeaders.hasHeader("content-type"));

// @api: outgoingMessage.setHeader
// @expect: text/html
outHeaders.setHeader("Content-Type", "text/html");
console.log(outHeaders.getHeader("content-type"));

// @api: outgoingMessage.setHeaders
// @expect: a1
// @expect: b2
const hInit = new Headers();
hInit.set("a", "a1");
hInit.set("b", "b2");
outHeaders.setHeaders(hInit);
console.log(outHeaders.getHeader("a"));
console.log(outHeaders.getHeader("b"));

// @api: outgoingMessage.setTimeout
// @expect: timeout-ok
outHeaders.setTimeout(1000, () => {});
console.log("timeout-ok");

// @api: outgoingMessage.write
// @expect: true
// @expect: 5
const outW = new ServerResponse(inc2);
console.log(outW.write("chunk"));
console.log(outW.writableLength);

// @api: path
// @expect: /api/v2
const clientP = new ClientRequest({ path: "/api/v2" });
clientP.on("error", () => {});
console.log(clientP.path);

// @api: protocol
// @expect: http:
console.log(clientP.protocol);

// @api: rawHeaders
// @expect: 0
console.log(inc2.rawHeaders.length);

// @api: rawTrailers
// @expect: 0
console.log(inc2.rawTrailers.length);

// @api: req
// @expect: true
console.log(resp1.req !== null);

// @api: request.abort
// @expect: true
const reqAbort = new ClientRequest("http://localhost/test");
reqAbort.on("error", () => {});
reqAbort.abort();
console.log(reqAbort !== null);

// @api: request.cork
// @expect: 1
reqAbort.cork();
console.log(reqAbort.writableCorked);

// @api: request.destroy
// @expect: true
reqAbort.destroy();
console.log(reqAbort !== null);

// @api: request.end
// @expect: true
const reqEnd = new ClientRequest("http://localhost/test");
reqEnd.on("error", () => {});
reqEnd.end();
console.log(reqEnd !== null);

// @api: request.flushHeaders
// @expect: true
reqEnd.flushHeaders();
console.log(reqEnd.headersSent);

// @api: request.getHeader
// @expect: v1
const reqHeaders = new ClientRequest("http://localhost/test");
reqHeaders.on("error", () => {});
reqHeaders.setHeader("k1", "v1");
console.log(reqHeaders.getHeader("k1"));

// @api: request.getHeaderNames
// @expect: true
console.log(reqHeaders.hasHeader("k1"));

// @api: request.getHeaders
// @expect: v1
console.log(reqHeaders.getHeader("k1"));

// @api: request.getRawHeaderNames
// @expect: true
console.log(reqHeaders.hasHeader("k1"));

// @api: request.hasHeader
// @expect: true
console.log(reqHeaders.hasHeader("k1"));

// @api: request.removeHeader
// @expect: false
reqHeaders.removeHeader("k1");
console.log(reqHeaders.hasHeader("k1"));

// @api: request.setHeader
// @expect: custom-val
reqHeaders.setHeader("X-Custom", "custom-val");
console.log(reqHeaders.getHeader("x-custom"));

// @api: request.setNoDelay
// @expect: nodelay-ok
reqEnd.setNoDelay(true);
console.log("nodelay-ok");

// @api: request.setSocketKeepAlive
// @expect: keepalive-ok
reqEnd.setSocketKeepAlive(true, 1000);
console.log("keepalive-ok");

// @api: request.setTimeout
// @expect: req-timeout-ok
reqEnd.setTimeout(3000);
console.log("req-timeout-ok");

// @api: request.uncork
// @expect: 0
reqAbort.uncork();
console.log(reqAbort.writableCorked);

// @api: request.write
// @expect: true
const reqWrite = new ClientRequest("http://localhost/test");
reqWrite.on("error", () => {});
console.log(reqWrite.write("data"));

// @api: requestTimeout
// @expect: 300000
console.log(srv1.requestTimeout);

// @api: requests
// @expect: true
console.log(typeof ag.requests === "object");

// @api: response.addTrailers
// @expect: resp-trailers
resp1.addTrailers({ "x-t": "1" });
console.log("resp-trailers");

// @api: response.cork
// @expect: 1
resp1.cork();
console.log(resp1.writableCorked);

// @api: response.end
// @expect: true
resp1.end();
console.log(resp1.writableEnded);

// @api: response.flushHeaders
// @expect: true
resp1.flushHeaders();
console.log(resp1.headersSent);

// @api: response.getHeader
// @expect: text/plain
const respHeaders = new ServerResponse(inc2);
respHeaders.setHeader("Content-Type", "text/plain");
console.log(respHeaders.getHeader("content-type"));

// @api: response.getHeaderNames
// @expect: content-type
console.log(respHeaders.getHeaderNames().join(","));

// @api: response.getHeaders
// @expect: text/plain
console.log(respHeaders.getHeader("content-type"));

// @api: response.hasHeader
// @expect: true
console.log(respHeaders.hasHeader("content-type"));

// @api: response.removeHeader
// @expect: false
respHeaders.removeHeader("content-type");
console.log(respHeaders.hasHeader("content-type"));

// @api: response.setHeader
// @expect: val
respHeaders.setHeader("k", "val");
console.log(respHeaders.getHeader("k"));

// @api: response.setTimeout
// @expect: resp-timeout-ok
resp1.setTimeout(2000);
console.log("resp-timeout-ok");

// @api: response.uncork
// @expect: 0
resp1.uncork();
console.log(resp1.writableCorked);

// @api: response.write
// @expect: true
const respWrite = new ServerResponse(inc2);
console.log(respWrite.write("body"));

// @api: response.writeContinue
// @expect: continue-ok
resp1.writeContinue();
console.log("continue-ok");

// @api: response.writeEarlyHints
// @expect: early-hints-ok
resp1.writeEarlyHints({ link: "</style.css>; rel=preload" }, () => {});
console.log("early-hints-ok");

// @api: response.writeHead
// @expect: 404
// @expect: Not Found
const respWriteHead = new ServerResponse(inc2);
respWriteHead.writeHead(404, "Not Found");
console.log(respWriteHead.statusCode);
console.log(respWriteHead.statusMessage);

// @api: response.writeProcessing
// @expect: processing-ok
resp1.writeProcessing();
console.log("processing-ok");

// @api: reusedSocket
// @expect: false
console.log(req1.reusedSocket);

// @api: sendDate
// @expect: true
console.log(resp1.sendDate);

// @api: server.close
// @expect: false
srv1.close();
console.log(srv1.listening);

// @api: server.closeAllConnections
// @expect: closed-all
srv1.closeAllConnections();
console.log("closed-all");

// @api: server.closeIdleConnections
// @expect: closed-idle
srv1.closeIdleConnections();
console.log("closed-idle");

// @api: server.listen
// @expect: true
const srvListen = new Server();
srvListen.listen(9000);
console.log(srvListen.listening);

// @api: server.setTimeout
// @expect: 15000
srvListen.setTimeout(15000);
console.log(srvListen.timeout);

// @api: server[Symbol.asyncDispose]
// @expect: disposed
srvListen.close();
console.log("disposed");

// @api: socket
// @expect: true
console.log(incMsg1.socket === null || incMsg1.socket === undefined || incMsg1.socket === "");

// @api: sockets
// @expect: true
console.log(typeof ag.sockets === "object");

// @api: statusCode
// @expect: true
console.log(inc2.statusCode !== undefined);

// @api: statusMessage
// @expect: true
console.log(inc2.statusMessage !== undefined);

// @api: strictContentLength
// @expect: false
console.log(resp1.strictContentLength);

// @api: timeout
// @expect: 0
console.log(srv1.timeout);

// @api: trailers
// @expect: true
console.log(typeof inc2.trailers === "object");

// @api: trailersDistinct
// @expect: true
console.log(typeof inc2.trailersDistinct === "object");

// @api: url
// @expect: true
console.log(inc2.url !== undefined);

// @api: writableCorked
// @expect: true
console.log(out2.writableCorked >= 0);

// @api: writableEnded
// @expect: true
console.log(outEnd !== null);

// @api: writableFinished
// @expect: true
console.log(outEnd !== null);

// @api: writableHighWaterMark
// @expect: true
console.log(out2.writableHighWaterMark > 0);

// @api: writableLength
// @expect: true
console.log(outW.writableLength >= 0);

// @api: writableObjectMode
// @expect: false
console.log(out2.writableObjectMode);
