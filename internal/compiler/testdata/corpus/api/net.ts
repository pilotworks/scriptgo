// ScriptGo Corpus: Net Standard Builtin APIs
// Consolidated test suite with inline assertions.

import {
    isIP,
    isIPv4,
    isIPv6,
    SocketAddress,
    Socket,
    Server,
    createServer
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
// @expect: open
// @expect: 0
// @expect: 0
// @expect: false
// @expect: false
// @expect: true
// @expect: 5000
// @expect: 0
const sock = new Socket();
console.log(sock.readyState);
console.log(sock.bytesRead);
console.log(sock.bytesWritten);
console.log(sock.connecting);
console.log(sock.destroyed);
console.log(sock.pending);
sock.setTimeout(5000);
console.log(sock.timeout);
console.log(sock.autoSelectFamilyAttemptedAddresses !== undefined ? sock.autoSelectFamilyAttemptedAddresses.length : 0);
sock.destroy();

// Server state is observed only after the listening callback, matching Node's
// asynchronous lifecycle and avoiding writes to an unconnected socket.
// @api: net.Server
// @api: net.createServer
// @api: server.listen
// @api: server.address
// @api: server.close
// @api: listening
// @api: server.getConnections
// @api: socket.connect
// @api: socket.write
// @api: socket.end
// @api: socket.destroy
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
srv.listen(9000, "127.0.0.1", () => {
    console.log(srv.listening);
    console.log(srv.address().port);
    srv.getConnections((_err: unknown, count: number) => console.log(count));
    const clientSock = new Socket();
    clientSock.connect(9000, "127.0.0.1", () => {
        console.log(clientSock.remotePort);
        console.log(clientSock.remoteAddress);
        console.log(clientSock.remoteFamily);
        console.log(clientSock.write("hello network"));
        clientSock.end(() => {
            console.log(clientSock.readyState !== "open");
            clientSock.destroy();
            srv.close(() => console.log(srv.listening));
        });
    });
});
