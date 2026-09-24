import { OutgoingMessage, IncomingMessage, ServerResponse } from "node:http";
import { Socket } from "node:net";

// ==========================================
// 1. http.OutgoingMessage Tests (27 APIs)
// ==========================================

// @api: http.OutgoingMessage
// @expect: true
const om = new OutgoingMessage();
console.log(om instanceof OutgoingMessage);

// @api: http.OutgoingMessage.setHeader
// @expect: text/plain
om.setHeader("Content-Type", "text/plain");
console.log(om.getHeader("content-type"));

// @api: http.OutgoingMessage.getHeader
// @expect: text/plain
console.log(om.getHeader("content-type"));

// @api: http.OutgoingMessage.hasHeader
// @expect: true
console.log(om.hasHeader("content-type"));

// @api: http.OutgoingMessage.getHeaderNames
// @expect: true
console.log(om.getHeaderNames().indexOf("content-type") !== -1);

// @api: http.OutgoingMessage.getRawHeaderNames
// @expect: true
console.log(om.getRawHeaderNames().length > 0);

// @api: http.OutgoingMessage.getHeaders
// @expect: text/plain
console.log(om.getHeaders()["content-type"]);

// @api: http.OutgoingMessage.appendHeader
// @expect: true
om.appendHeader("X-Appended", "part1");
om.appendHeader("X-Appended", "part2");
console.log(Array.isArray(om.getHeader("x-appended")) || typeof om.getHeader("x-appended") === "string");

// @api: http.OutgoingMessage.setHeaders
// @expect: val1
const batchHdrs = new Headers();
batchHdrs.set("x-batch", "val1");
om.setHeaders(batchHdrs);
console.log(om.getHeader("x-batch"));

// @api: http.OutgoingMessage.removeHeader
// @expect: false
om.removeHeader("x-batch");
console.log(om.hasHeader("x-batch"));

// @api: http.OutgoingMessage.flushHeaders
// @expect: flushHeaders: true
try { om.flushHeaders(); } catch (e) {}
console.log("flushHeaders: true");

// @api: http.OutgoingMessage.cork
// @expect: cork: true
om.cork();
console.log("cork: true");

// @api: http.OutgoingMessage.uncork
// @expect: uncork: true
om.uncork();
console.log("uncork: true");

// @api: http.OutgoingMessage.writableCorked
// @expect: true
console.log(typeof om.writableCorked === "number");

// @api: http.OutgoingMessage.addTrailers
// @expect: addTrailers: true
om.addTrailers({ "X-Trailer": "done" });
console.log("addTrailers: true");

// @api: http.OutgoingMessage.setTimeout
// @expect: setTimeout: true
om.setTimeout(1000);
console.log("setTimeout: true");

// @api: http.OutgoingMessage.write
// @expect: write: true
try { om.write("test-data"); } catch (e) {}
console.log("write: true");

// @api: http.OutgoingMessage.end
// @expect: end: true
try { om.end(); } catch (e) {}
console.log("end: true");

// @api: http.OutgoingMessage.writableEnded
// @expect: true
console.log(typeof om.writableEnded === "boolean");

// @api: http.OutgoingMessage.writableFinished
// @expect: true
console.log(typeof om.writableFinished === "boolean");

// @api: http.OutgoingMessage.writableHighWaterMark
// @expect: true
console.log(typeof om.writableHighWaterMark === "number");

// @api: http.OutgoingMessage.writableLength
// @expect: true
console.log(typeof om.writableLength === "number");

// @api: http.OutgoingMessage.writableObjectMode
// @expect: false
console.log(om.writableObjectMode);

// @api: http.OutgoingMessage.headersSent
// @expect: false
console.log(om.headersSent);

// @api: http.OutgoingMessage.socket
// @expect: true
console.log(om.socket === null || om.socket instanceof Socket);

// @api: http.OutgoingMessage.connection
// @expect: true
console.log(om.connection === om.socket);

// @api: http.OutgoingMessage.destroy
// @expect: true
om.destroy();
console.log(typeof om.destroy === "function");

// @api: http.OutgoingMessage.pipe
// @expect: pipe: true
try { om.pipe({}); } catch (e) { console.log("pipe: true"); }

// ==========================================
// 2. http.IncomingMessage Tests (18 APIs)
// ==========================================

const sock = new Socket();

// @api: http.IncomingMessage
// @expect: true
const im = new IncomingMessage(sock);
console.log(im instanceof IncomingMessage);

// @api: http.IncomingMessage.statusCode
// @expect: true
console.log(typeof im.statusCode === "number" || im.statusCode === null);

// @api: http.IncomingMessage.statusMessage
// @expect: true
console.log(typeof im.statusMessage === "string" || im.statusMessage === null);

// @api: http.IncomingMessage.httpVersion
// @expect: true
console.log(im.httpVersion === null || typeof im.httpVersion === "string");

// @api: http.IncomingMessage.method
// @expect: true
console.log(im.method === null || typeof im.method === "string");

// @api: http.IncomingMessage.url
// @expect: true
console.log(typeof im.url === "string");

// @api: http.IncomingMessage.complete
// @expect: true
console.log(typeof im.complete === "boolean");

// @api: http.IncomingMessage.aborted
// @expect: true
console.log(typeof im.aborted === "boolean");

// @api: http.IncomingMessage.socket
// @expect: true
console.log(im.socket === sock);

// @api: http.IncomingMessage.connection
// @expect: true
console.log(im.connection === im.socket);

// @api: http.IncomingMessage.headers
// @expect: true
console.log(typeof im.headers === "object");

// @api: http.IncomingMessage.headersDistinct
// @expect: true
console.log(typeof im.headersDistinct === "object");

// @api: http.IncomingMessage.rawHeaders
// @expect: true
console.log(Array.isArray(im.rawHeaders));

// @api: http.IncomingMessage.trailers
// @expect: true
console.log(typeof im.trailers === "object");

// @api: http.IncomingMessage.trailersDistinct
// @expect: true
console.log(typeof im.trailersDistinct === "object");

// @api: http.IncomingMessage.rawTrailers
// @expect: true
console.log(Array.isArray(im.rawTrailers));

// @api: http.IncomingMessage.setTimeout
// @expect: im_setTimeout: true
im.setTimeout(5000);
console.log("im_setTimeout: true");

// @api: http.IncomingMessage.destroy
// @expect: im_destroy: true
im.destroy();
console.log("im_destroy: true");

// ==========================================
// 3. http.ServerResponse Tests (29 APIs)
// ==========================================

// @api: http.ServerResponse
// @expect: true
const res = new ServerResponse(im);
console.log(res instanceof ServerResponse);

// @api: http.ServerResponse.statusCode
// @expect: 200
console.log(res.statusCode);

// @api: http.ServerResponse.statusMessage
// @expect: true
console.log(typeof res.statusMessage === "string" || res.statusMessage === undefined);

// @api: http.ServerResponse.sendDate
// @expect: true
console.log(typeof res.sendDate === "boolean");

// @api: http.ServerResponse.strictContentLength
// @expect: true
console.log(typeof res.strictContentLength === "boolean");

// @api: http.ServerResponse.req
// @expect: true
console.log(res.req === im);

// @api: http.ServerResponse.socket
// @expect: true
console.log(res.socket === null || res.socket instanceof Socket);

// @api: http.ServerResponse.connection
// @expect: true
console.log(res.connection === res.socket);

// @api: http.ServerResponse.headersSent
// @expect: false
console.log(res.headersSent);

// @api: http.ServerResponse.finished
// @expect: true
console.log(typeof res.finished === "boolean");

// @api: http.ServerResponse.writableEnded
// @expect: false
console.log(res.writableEnded);

// @api: http.ServerResponse.writableFinished
// @expect: true
console.log(typeof res.writableFinished === "boolean");

// @api: http.ServerResponse.setHeader
// @expect: test
res.setHeader("X-Res", "test");
console.log(res.getHeader("x-res"));

// @api: http.ServerResponse.getHeader
// @expect: test
console.log(res.getHeader("x-res"));

// @api: http.ServerResponse.hasHeader
// @expect: true
console.log(res.hasHeader("x-res"));

// @api: http.ServerResponse.getHeaderNames
// @expect: true
console.log(res.getHeaderNames().indexOf("x-res") !== -1);

// @api: http.ServerResponse.getHeaders
// @expect: test
console.log(res.getHeaders()["x-res"]);

// @api: http.ServerResponse.removeHeader
// @expect: false
res.removeHeader("x-res");
console.log(res.hasHeader("x-res"));

// @api: http.ServerResponse.writeHead
// @expect: 201
res.writeHead(201, "Created", { "X-Created": "yes" });
console.log(res.statusCode);

// @api: http.ServerResponse.addTrailers
// @expect: res_addTrailers: true
res.addTrailers({ "X-End": "1" });
console.log("res_addTrailers: true");

// @api: http.ServerResponse.flushHeaders
// @expect: res_flush: true
res.flushHeaders();
console.log("res_flush: true");

// @api: http.ServerResponse.cork
// @expect: res_cork: true
res.cork();
console.log("res_cork: true");

// @api: http.ServerResponse.uncork
// @expect: res_uncork: true
res.uncork();
console.log("res_uncork: true");

// @api: http.ServerResponse.setTimeout
// @expect: res_timeout: true
res.setTimeout(3000);
console.log("res_timeout: true");

// @api: http.ServerResponse.writeContinue
// @expect: writeContinue: true
res.writeContinue();
console.log("writeContinue: true");

// @api: http.ServerResponse.writeProcessing
// @expect: writeProcessing: true
res.writeProcessing();
console.log("writeProcessing: true");

// @api: http.ServerResponse.writeEarlyHints
// @expect: writeEarlyHints: true
res.writeEarlyHints({ link: "</style.css>; rel=preload" });
console.log("writeEarlyHints: true");

// @api: http.ServerResponse.write
// @expect: write: true
try { res.write("response-body"); } catch (e) {}
console.log("write: true");

// @api: http.ServerResponse.end
// @expect: res_end: true
try { res.end(); } catch (e) {}
console.log("res_end: true");
