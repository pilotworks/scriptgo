import {
    constants,
    sensitiveHeaders,
    Http2Session,
    ServerHttp2Session,
    ClientHttp2Session,
    Http2Stream,
    ClientHttp2Stream,
    ServerHttp2Stream,
    Http2Server,
    Http2SecureServer,
    Http2ServerRequest,
    Http2ServerResponse,
    createServer,
    createSecureServer,
    connect,
    getDefaultSettings,
    getPackedSettings,
    getUnpackedSettings,
    performServerHandshake,
} from "node:http2";
import { Duplex } from "node:stream";

// @api: http2.constants
// @api: http2.sensitiveHeaders
// @expect: h2_props: 0 symbol
console.log("h2_props: " + constants.NGHTTP2_SESSION_SERVER + " " + typeof sensitiveHeaders);

// @api: http2.getDefaultSettings
// @api: http2.getPackedSettings
// @api: http2.getUnpackedSettings
// @expect: h2_settings: 4096 42 4096
const defSettings = getDefaultSettings();
const packed = getPackedSettings(defSettings);
const unpacked = getUnpackedSettings(new Uint8Array(0));
console.log("h2_settings: " + defSettings.headerTableSize + " " + packed.length + " " + unpacked.headerTableSize);

// @api: http2.http2.Http2Session
// @api: http2.Http2Session
// @api: new http2.Http2Session
// @api: Http2Session.alpnProtocol
// @api: Http2Session.closed
// @api: Http2Session.connecting
// @api: Http2Session.destroyed
// @api: Http2Session.encrypted
// @api: Http2Session.localSettings
// @api: Http2Session.originSet
// @api: Http2Session.pendingSettingsAck
// @api: Http2Session.remoteSettings
// @api: Http2Session.socket
// @api: Http2Session.state
// @api: Http2Session.type
// @api: Http2Session.close
// @api: Http2Session.destroy
// @api: Http2Session.goaway
// @api: Http2Session.ping
// @api: Http2Session.ref
// @api: Http2Session.unref
// @api: Http2Session.setLocalWindowSize
// @api: Http2Session.setTimeout
// @api: Http2Session.settings
// @expect: h2_session: h2 false false false true 0 true true
const session = new Http2Session();
session.goaway(0);
session.ping(new Uint8Array(8), (err, dur, buf) => {});
session.ref();
session.unref();
session.setLocalWindowSize(65535);
session.setTimeout(1000, () => {});
session.settings({ headerTableSize: 4096 });
console.log("h2_session: " + session.alpnProtocol + " " + session.closed + " " + session.connecting + " " + session.destroyed + " " + session.encrypted + " " + session.type + " " + (typeof session.socket === "object") + " " + (typeof session.localSettings === "object"));
session.close();
session.destroy();

// @api: http2.http2.ServerHttp2Session
// @api: http2.ServerHttp2Session
// @api: new http2.ServerHttp2Session
// @api: ServerHttp2Session.altsvc
// @api: ServerHttp2Session.origin
// @api: http2.performServerHandshake
// @expect: h2_serverSession: true
const sSession = performServerHandshake(new Duplex());
sSession.altsvc("h2", "origin");
sSession.origin("https://example.com");
console.log("h2_serverSession: " + (sSession instanceof ServerHttp2Session));

// @api: http2.http2.ClientHttp2Session
// @api: http2.ClientHttp2Session
// @api: new http2.ClientHttp2Session
// @api: http2.connect
// @api: ClientHttp2Session.request
// @expect: h2_clientSession: true true
const cSession = connect("https://localhost:8443");
const clientStream = cSession.request({ ":path": "/" });
console.log("h2_clientSession: " + (cSession instanceof ClientHttp2Session) + " " + (clientStream instanceof ClientHttp2Stream));

// @api: http2.http2.Http2Stream
// @api: http2.Http2Stream
// @api: new http2.Http2Stream
// @api: Http2Stream.aborted
// @api: Http2Stream.bufferSize
// @api: Http2Stream.closed
// @api: Http2Stream.destroyed
// @api: Http2Stream.endAfterHeaders
// @api: Http2Stream.id
// @api: Http2Stream.pending
// @api: Http2Stream.rstCode
// @api: Http2Stream.sentHeaders
// @api: Http2Stream.sentInfoHeaders
// @api: Http2Stream.sentTrailers
// @api: Http2Stream.session
// @api: Http2Stream.state
// @api: Http2Stream.close
// @api: Http2Stream.priority
// @api: Http2Stream.setTimeout
// @api: Http2Stream.sendTrailers
// @expect: h2_stream: false 0 1 0 true true
const st = new Http2Stream();
st.priority({});
st.setTimeout(1000, () => {});
st.sendTrailers({ trailer: "val" });
console.log("h2_stream: " + st.aborted + " " + st.bufferSize + " " + st.id + " " + st.rstCode + " " + (typeof st.session === "object") + " " + (typeof st.sentHeaders === "object"));
st.close();

// @api: http2.http2.ClientHttp2Stream
// @api: http2.ClientHttp2Stream
// @api: new http2.ClientHttp2Stream
// @expect: h2_clientStream: true
const cs = new ClientHttp2Stream();
console.log("h2_clientStream: " + (cs instanceof ClientHttp2Stream));

// @api: http2.http2.ServerHttp2Stream
// @api: http2.ServerHttp2Stream
// @api: new http2.ServerHttp2Stream
// @api: ServerHttp2Stream.headersSent
// @api: ServerHttp2Stream.pushAllowed
// @api: ServerHttp2Stream.additionalHeaders
// @api: ServerHttp2Stream.pushStream
// @api: ServerHttp2Stream.respond
// @api: ServerHttp2Stream.respondWithFD
// @api: ServerHttp2Stream.respondWithFile
// @expect: h2_serverStream: true true
const sst = new ServerHttp2Stream();
sst.additionalHeaders({ "x-custom": "1" });
sst.pushStream({ ":path": "/sub" }, {}, (err, pStream, headers) => {});
sst.respond({ ":status": 200 });
sst.respondWithFD(1, { ":status": 200 });
sst.respondWithFile("test.txt", { ":status": 200 });
console.log("h2_serverStream: " + sst.headersSent + " " + sst.pushAllowed);

// @api: http2.http2.Http2Server
// @api: http2.Http2Server
// @api: new http2.Http2Server
// @api: http2.createServer
// @api: Http2Server.timeout
// @api: Http2Server.setTimeout
// @api: Http2Server.updateSettings
// @api: Http2Server.close
// @api: Http2Server.[Symbol.asyncDispose]
// @expect: h2_server: 5000 true
const srv = createServer();
srv.setTimeout(5000);
srv.updateSettings({ headerTableSize: 4096 });
console.log("h2_server: " + srv.timeout + " " + (srv instanceof Http2Server));
srv.close();

// @api: http2.http2.Http2SecureServer
// @api: http2.Http2SecureServer
// @api: new http2.Http2SecureServer
// @api: http2.createSecureServer
// @api: Http2SecureServer.timeout
// @api: Http2SecureServer.setTimeout
// @api: Http2SecureServer.updateSettings
// @api: Http2SecureServer.close
// @expect: h2_secServer: 5000 true
const secSrv = createSecureServer();
secSrv.setTimeout(5000);
secSrv.updateSettings({ headerTableSize: 4096 });
console.log("h2_secServer: " + secSrv.timeout + " " + (secSrv instanceof Http2SecureServer));
secSrv.close();

// @api: http2.http2.http2.Http2ServerRequest
// @api: http2.Http2ServerRequest
// @api: new http2.Http2ServerRequest
// @api: http2.Http2ServerRequest.aborted
// @api: http2.Http2ServerRequest.authority
// @api: http2.Http2ServerRequest.complete
// @api: http2.Http2ServerRequest.connection
// @api: http2.Http2ServerRequest.headers
// @api: http2.Http2ServerRequest.httpVersion
// @api: http2.Http2ServerRequest.method
// @api: http2.Http2ServerRequest.rawHeaders
// @api: http2.Http2ServerRequest.rawTrailers
// @api: http2.Http2ServerRequest.scheme
// @api: http2.Http2ServerRequest.socket
// @api: http2.Http2ServerRequest.stream
// @api: http2.Http2ServerRequest.trailers
// @api: http2.Http2ServerRequest.url
// @api: http2.Http2ServerRequest.setTimeout
// @api: http2.Http2ServerRequest.destroy
// @expect: h2_req: false true 2.0 GET https / true true
const req = new Http2ServerRequest();
req.setTimeout(1000);
console.log("h2_req: " + req.aborted + " " + req.complete + " " + req.httpVersion + " " + req.method + " " + req.scheme + " " + req.url + " " + (typeof req.connection === "object") + " " + (typeof req.stream === "object"));
req.destroy();

// @api: http2.http2.http2.Http2ServerResponse
// @api: http2.Http2ServerResponse
// @api: new http2.Http2ServerResponse
// @api: http2.Http2ServerResponse.connection
// @api: http2.Http2ServerResponse.finished
// @api: http2.Http2ServerResponse.headersSent
// @api: http2.Http2ServerResponse.req
// @api: http2.Http2ServerResponse.sendDate
// @api: http2.Http2ServerResponse.socket
// @api: http2.Http2ServerResponse.statusCode
// @api: http2.Http2ServerResponse.statusMessage
// @api: http2.Http2ServerResponse.stream
// @api: http2.Http2ServerResponse.writableEnded
// @api: http2.Http2ServerResponse.addTrailers
// @api: http2.Http2ServerResponse.appendHeader
// @api: http2.Http2ServerResponse.createPushResponse
// @api: http2.Http2ServerResponse.end
// @api: http2.Http2ServerResponse.getHeader
// @api: http2.Http2ServerResponse.getHeaderNames
// @api: http2.Http2ServerResponse.getHeaders
// @api: http2.Http2ServerResponse.hasHeader
// @api: http2.Http2ServerResponse.removeHeader
// @api: http2.Http2ServerResponse.setHeader
// @api: http2.Http2ServerResponse.setTimeout
// @api: http2.Http2ServerResponse.write
// @api: http2.Http2ServerResponse.writeContinue
// @api: http2.Http2ServerResponse.writeEarlyHints
// @api: http2.Http2ServerResponse.writeHead
// @expect: h2_res: 200 OK val true true true
const res = new Http2ServerResponse();
res.setHeader("x-test", "val");
res.appendHeader("x-test2", "val2");
res.addTrailers({ "x-trailer": "1" });
res.writeContinue();
res.writeEarlyHints({ link: "preload" });
res.writeHead(200);
res.write("data");
const hVal = res.getHeader("x-test");
const hasH = res.hasHeader("x-test");
const hNames = res.getHeaderNames();
const allH = res.getHeaders();
res.removeHeader("x-test2");
res.createPushResponse({ ":path": "/res" }, (err, pushRes) => {});
res.setTimeout(1000);
res.end();
console.log("h2_res: " + res.statusCode + " " + res.statusMessage + " " + hVal + " " + hasH + " " + res.finished + " " + res.writableEnded);

// @expect: h2_async: true
const runH2Async = async () => {
    await srv[Symbol.asyncDispose]();
    console.log("h2_async: true");
};
runH2Async();
