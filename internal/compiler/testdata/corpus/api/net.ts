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
import * as net from "node:net";

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

// @api: net.BlockList
// @api: BlockList.isBlockList
// @api: blockList.addAddress
// @api: blockList.addRange
// @api: blockList.addSubnet
// @api: blockList.check
// @api: blockList.toJSON
// @api: rules
// @expect: true
// @expect: true
// @expect: false
// @expect: 3
const bl = new BlockList();
console.log(BlockList.isBlockList(bl));
bl.addAddress("10.0.0.1");
bl.addRange("10.0.0.10", "10.0.0.20");
bl.addSubnet("192.168.0.0", 24);
console.log(bl.check("10.0.0.1"));
console.log(bl.check("10.0.0.2"));
console.log(bl.rules.length);

// @api: blockList.fromJSON
// @expect: 4
bl.fromJSON(["addr:ipv4:172.16.0.1"]);
const jsonRules = bl.toJSON();
console.log(jsonRules.length);

// @api: net.Socket
// @api: socket.address
// @api: socket.setTimeout
// @api: socket.setNoDelay
// @api: socket.setKeepAlive
// @api: socket.setEncoding
// @api: socket.pause
// @api: socket.resume
// @api: socket.ref
// @api: socket.unref
// @api: readyState
// @api: localAddress
// @api: localFamily
// @api: localPort
// @api: remoteAddress
// @api: remoteFamily
// @api: remotePort
// @api: bufferSize
// @api: bytesRead
// @api: bytesWritten
// @api: connecting
// @api: destroyed
// @api: pending
// @api: timeout
// @api: autoSelectFamilyAttemptedAddresses
// @expect: open
// @expect: 127.0.0.1
// @expect: IPv4
// @expect: 0
// @expect: 0
// @expect: 0
// @expect: 0
// @expect: false
// @expect: false
// @expect: false
// @expect: 5000
// @expect: 0
// @expect: 127.0.0.1
const sock = new Socket();
console.log(sock.readyState);
console.log(sock.localAddress);
console.log(sock.localFamily);
console.log(sock.localPort);
console.log(sock.bufferSize);
console.log(sock.bytesRead);
console.log(sock.bytesWritten);
console.log(sock.connecting);
console.log(sock.destroyed);
console.log(sock.pending);
sock.setTimeout(5000);
console.log(sock.timeout);
console.log(sock.autoSelectFamilyAttemptedAddresses.length);
sock.setNoDelay(true);
sock.setKeepAlive(true, 1000);
sock.setEncoding("utf8");
sock.pause();
sock.resume();
sock.ref();
sock.unref();
const sockAddr = sock.address();
console.log(sockAddr.address);

// @api: socket.write
// @api: socket.end
// @api: socket.destroy
// @api: socket.destroySoon
// @api: socket.resetAndDestroy
// @expect: true
// @expect: closed
console.log(sock.write("hello network"));
sock.end();
console.log(sock.readyState);
sock.destroySoon();
sock.resetAndDestroy();

// @api: net.Server
// @api: net.createServer
// @api: server.listen
// @api: server.address
// @api: server.ref
// @api: server.unref
// @api: server.close
// @api: listening
// @api: maxConnections
// @api: dropMaxConnection
// @api: server[Symbol.asyncDispose]()
// @expect: true
// @expect: 1000
// @expect: false
// @expect: 8080
// @expect: false
const srv = createServer();
srv.listen(9000);
console.log(srv.listening);
console.log(srv.maxConnections);
console.log(srv.dropMaxConnection);
const srvAddr = srv.address();
console.log(srvAddr.port);
srv.ref();
srv.unref();
srv.close();
console.log(srv.listening);
srv.asyncDispose();

// @api: server.getConnections
// @expect: 0
srv.getConnections((err: unknown, count: number) => {
    console.log(count);
});

// @api: net.connect
// @api: net.createConnection
// @api: socket.connect
// @expect: 8080
// @expect: localhost
// @expect: IPv4
const clientSock = connect(8080, "localhost");
console.log(clientSock.remotePort);
console.log(clientSock.remoteAddress);
console.log(clientSock.remoteFamily);

// @api: net.getDefaultAutoSelectFamily
// @api: net.setDefaultAutoSelectFamily
// @expect: true
setDefaultAutoSelectFamily(true);
console.log(getDefaultAutoSelectFamily());

// @api: net.getDefaultAutoSelectFamilyAttemptTimeout
// @api: net.setDefaultAutoSelectFamilyAttemptTimeout
// @expect: 250
setDefaultAutoSelectFamilyAttemptTimeout(250);
console.log(getDefaultAutoSelectFamilyAttemptTimeout());
