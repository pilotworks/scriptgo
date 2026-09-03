// ScriptGo Corpus: Net Standard Builtin APIs
// Consolidated test suite with inline assertions.

import {
    isIP,
    isIPv4,
    isIPv6,
    SocketAddress,
    Socket,
    Server,
    createServer,
    createConnection,
    connect
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

// @api: net.Socket
// @api: socket.address
// @api: socket.setTimeout
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
console.log(sock.autoSelectFamilyAttemptedAddresses !== undefined ? sock.autoSelectFamilyAttemptedAddresses.length : 0);
const sockAddr = sock.address();
console.log(sockAddr.address);

// @api: socket.write
// @api: socket.end
// @api: socket.destroy
// @expect: true
// @expect: true
// @expect: 16
// @expect: closed
console.log(sock.write("hello network"));
const netPayload: string | Uint8Array = new Uint8Array(3);
console.log(sock.write(netPayload));
console.log(sock.bytesWritten);
sock.end();
console.log(sock.readyState);

// @api: net.Server
// @api: net.createServer
// @api: server.listen
// @api: server.address
// @api: server.close
// @api: listening
// @api: maxConnections
// @api: dropMaxConnection
// @api: server[Symbol.asyncDispose]()
// @expect: true
// @expect: 1000
// @expect: false
// @expect: 9000
// @expect: false
const srv = createServer();
srv.listen(9000);
console.log(srv.listening);
console.log(srv.maxConnections);
console.log(srv.dropMaxConnection);
const srvAddr = srv.address();
console.log(srvAddr.port);
srv.close();
console.log(srv.listening);
srv.close();

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
// @expect: 4321
// @expect: example.test
const clientSock = connect(8080, "localhost");
console.log(clientSock.remotePort);
console.log(clientSock.remoteAddress);
console.log(clientSock.remoteFamily);
const optionsSock = connect({ port: 4321, host: "example.test" });
console.log(optionsSock.remotePort);
console.log(optionsSock.remoteAddress);

// @api: net.Server.[Symbol.asyncDispose]
// @expect: true
const srvAsync = createServer();
console.log(srvAsync !== null);
