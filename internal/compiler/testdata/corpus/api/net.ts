// ScriptGo Corpus: Net Standard Builtin APIs
// Consolidated test suite with inline assertions.

import {
    isIP,
    isIPv4,
    isIPv6,
    SocketAddress,
    BlockList,
    Socket,
    Server,
    createServer,
    createConnection,
    connect,
    getDefaultAutoSelectFamily,
    setDefaultAutoSelectFamily,
    getDefaultAutoSelectFamilyAttemptTimeout,
    setDefaultAutoSelectFamilyAttemptTimeout
} from "node:net";

// @api: net.isIPv4
// @expect: true
// @expect: false
console.log(isIPv4("127.0.0.1"));
console.log(isIPv4("256.0.0.1"));

// @api: net.isIPv6
// @expect: true
// @expect: false
console.log(isIPv6("2001:0db8:85a3:0000:0000:8a2e:0370:7334"));
console.log(isIPv6("127.0.0.1"));

// @api: net.isIP
// @expect: 4
// @expect: 6
// @expect: 0
console.log(isIP("192.168.1.1"));
console.log(isIP("fe80::1"));
console.log(isIP("invalid_ip"));

// @api: net.getDefaultAutoSelectFamily
// @api: net.setDefaultAutoSelectFamily
// @api: net.getDefaultAutoSelectFamilyAttemptTimeout
// @api: net.setDefaultAutoSelectFamilyAttemptTimeout
// @expect: true
// @expect: true
// @expect: 250
// @expect: 300
console.log(getDefaultAutoSelectFamily());
setDefaultAutoSelectFamily(true);
console.log(getDefaultAutoSelectFamily());
setDefaultAutoSelectFamily(false);
console.log(getDefaultAutoSelectFamilyAttemptTimeout());
setDefaultAutoSelectFamilyAttemptTimeout(300);
console.log(getDefaultAutoSelectFamilyAttemptTimeout());

// @api: net.BlockList
// @api: net.net.BlockList
// @api: new net.BlockList
// @api: BlockList.addAddress
// @api: BlockList.addRange
// @api: BlockList.addSubnet
// @api: BlockList.check
// @api: BlockList.rules
// @api: BlockList.toJSON
// @api: BlockList.fromJSON
// @api: BlockList.isBlockList
// @expect: true
// @expect: true
// @expect: true
// @expect: false
// @expect: 3
// @expect: 3
// @expect: true
// @expect: true
const bl = new BlockList();
bl.addAddress("127.0.0.1", "ipv4");
bl.addRange("192.168.1.1", "192.168.1.10", "ipv4");
bl.addSubnet("10.0.0.0", 24, "ipv4");
console.log(bl.check("127.0.0.1"));
console.log(bl.check("192.168.1.5"));
console.log(bl.check("10.0.0.50"));
console.log(bl.check("8.8.8.8"));
console.log(bl.rules.length);
console.log(bl.toJSON().length);
console.log(BlockList.isBlockList(bl));
const bl2 = new BlockList();
bl2.fromJSON(bl.toJSON());
console.log(bl2.check("10.0.0.1"));

// @api: net.SocketAddress
// @api: address
// @api: port
// @api: family
// @api: flowlabel
// @expect: 127.0.0.1
// @expect: 8080
// @expect: ipv4
// @expect: 0
const sa = new SocketAddress({ address: "127.0.0.1", port: 8080 });
console.log(sa.address);
console.log(sa.port);
console.log(sa.family);
console.log(sa.flowlabel);

// @api: SocketAddress.parse
// @expect: 192.168.1.5
// @expect: 3000
const parsedSa = SocketAddress.parse("192.168.1.5:3000");
console.log(parsedSa.address);
console.log(parsedSa.port);

// @api: net.Socket
// @api: socket.address
// @api: socket.setTimeout
// @api: readyState
// @api: bytesRead
// @api: bytesWritten
// @api: connecting
// @api: destroyed
// @api: pending
// @api: timeout
// @api: autoSelectFamilyAttemptedAddresses
// @api: socket.pause
// @api: socket.resume
// @api: socket.ref
// @api: socket.unref
// @api: socket.setEncoding
// @api: socket.bufferSize
// @api: socket.localAddress
// @api: socket.localPort
// @api: socket.localFamily
// @api: socket.remoteAddress
// @api: socket.remotePort
// @api: socket.remoteFamily
// @api: socket.destroySoon
// @api: socket.resetAndDestroy
// @expect: open
// @expect: 0
// @expect: 0
// @expect: false
// @expect: false
// @expect: true
// @expect: 5000
// @expect: 0
// @expect: true
// @expect: true
// @expect: true
// @expect: true
// @expect: true
// @expect: undefined
// @expect: undefined
// @expect: undefined
// @expect: undefined
// @expect: undefined
// @expect: undefined
// @expect: undefined
// @expect: true
// @expect: true
const sock = new Socket();
sock.on("error", () => {});
console.log(sock.readyState);
console.log(sock.bytesRead);
console.log(sock.bytesWritten);
console.log(sock.connecting);
console.log(sock.destroyed);
console.log(sock.pending);
sock.setTimeout(5000);
console.log(sock.timeout);
console.log(sock.autoSelectFamilyAttemptedAddresses !== undefined ? sock.autoSelectFamilyAttemptedAddresses.length : 0);
console.log(sock.pause() === sock);
console.log(sock.resume() === sock);
console.log(sock.ref() === sock);
console.log(sock.unref() === sock);
console.log(sock.setEncoding("utf8") === sock);
console.log(sock.bufferSize);
console.log(sock.localAddress);
console.log(sock.localPort);
console.log(sock.localFamily);
console.log(sock.remoteAddress);
console.log(sock.remotePort);
console.log(sock.remoteFamily);
console.log(sock.destroySoon() === undefined);
console.log(sock.resetAndDestroy() === sock);

// @api: socket.setNoDelay
// @api: socket.setKeepAlive
// @expect: true
// @expect: true
const sock2 = new Socket();
console.log(sock2.setNoDelay(true) === sock2);
console.log(sock2.setKeepAlive(true, 1000) === sock2);
sock2.destroy();

// @api: net.connect
// @api: net.createConnection
// @expect: true
// @expect: true
const c1 = connect(9001, "127.0.0.1");
console.log(c1 !== null);
c1.destroy();
const c2 = createConnection(9001, "127.0.0.1");
console.log(c2 !== null);
c2.destroy();

// Server state is observed only after the listening callback, matching Node's
// asynchronous lifecycle and avoiding writes to an unconnected socket.
// @api: net.Server
// @api: net.createServer
// @api: server.listen
// @api: server.address
// @api: server.close
// @api: listening
// @api: server.getConnections
// @api: server.ref
// @api: server.unref
// @api: server.maxConnections
// @api: server.dropMaxConnection
// @api: server.[Symbol.asyncDispose]
// @api: socket.connect
// @api: socket.write
// @api: socket.end
// @api: socket.destroy
// @expect: true
// @expect: true
// @expect: undefined
// @expect: undefined
// @expect: true
// @expect: 9000
// @expect: 0
// @expect: 9000
// @expect: 127.0.0.1
// @expect: IPv4
// @expect: true
// @expect: true
// @expect: false
const srv = createServer((connection: Socket) => {
    connection.on("data", () => {});
});
console.log(srv.ref() === srv);
console.log(srv.unref() === srv);
console.log(srv.maxConnections);
console.log(srv.dropMaxConnection);
srv[Symbol.asyncDispose]();

const srv2 = createServer((s: Socket) => {
    s.on("data", () => {});
    s.on("end", () => { s.end(); });
});
srv2.listen(9000, "127.0.0.1", () => {
    console.log(srv2.listening);
    const addr = srv2.address() as { port: number, family: string, address: string };
    console.log(addr.port);
    srv2.getConnections((_err: unknown, count: number) => console.log(count));
    const clientSock = new Socket();
    clientSock.connect(9000, "127.0.0.1", () => {
        console.log(clientSock.remotePort);
        console.log(clientSock.remoteAddress);
        console.log(clientSock.remoteFamily);
        console.log(clientSock.write("hello network"));
        clientSock.end(() => {
            console.log(clientSock.readyState !== "open");
            clientSock.destroy();
            srv2.close(() => console.log(srv2.listening));
        });
    });
});
