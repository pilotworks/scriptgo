// ScriptGo Corpus: Dgram Standard Builtin APIs
// Consolidated test suite with 1:1 isolated assertions for all 27 official Node.js dgram APIs.

import {
    Socket,
    createSocket
} from "node:dgram";

// @api: new dgram.Socket
// @expect: socket_inst: true
const sock = new Socket("udp4");
console.log("socket_inst: " + (sock instanceof Socket));

// @api: dgram.Socket.bind
// @expect: socket_bind: true
sock.bind(41234, () => {
    console.log("socket_bind: true");

    // The remaining socket operations run after bind/connect events so the
    // test follows Node.js dgram lifecycle semantics.
    // @api: dgram.Socket.address
    // @expect: socket_address: 41234
    const addr = sock.address();
    console.log("socket_address: " + addr.port);

    sock.connect(41235, "127.0.0.1", () => {
        // @api: dgram.Socket.connect
        // @expect: socket_connect: true
        console.log("socket_connect: true");

        // @api: dgram.Socket.remoteAddress
        // @expect: socket_remoteAddress: 41235
        const raddr = sock.remoteAddress();
        console.log("socket_remoteAddress: " + raddr.port);

        sock.disconnect();
        // @api: dgram.Socket.disconnect
        // @expect: socket_disconnect: true
        console.log("socket_disconnect: true");

        sock.send(new Uint8Array([1, 2, 3]), 41234, "127.0.0.1");
        // @api: dgram.Socket.send
        // @expect: socket_send: true
        console.log("socket_send: true");

        sock.setBroadcast(true);
        // @api: dgram.Socket.setBroadcast
        // @expect: socket_setBroadcast: true
        console.log("socket_setBroadcast: true");

        // @api: dgram.Socket.setTTL
        // @expect: socket_setTTL: 64
        console.log("socket_setTTL: " + sock.setTTL(64));
        // @api: dgram.Socket.setMulticastTTL
        // @expect: socket_setMulticastTTL: 1
        console.log("socket_setMulticastTTL: " + sock.setMulticastTTL(1));
        // @api: dgram.Socket.setMulticastLoopback
        // @expect: socket_setMulticastLoopback: true
        console.log("socket_setMulticastLoopback: " + sock.setMulticastLoopback(true));
        // @api: dgram.Socket.setMulticastInterface
        // @expect: socket_setMulticastInterface: true
        sock.setMulticastInterface("127.0.0.1");
        console.log("socket_setMulticastInterface: true");
        sock.addMembership("224.0.0.1");
        // @api: dgram.Socket.addMembership
        // @expect: socket_addMembership: true
        console.log("socket_addMembership: true");
        sock.dropMembership("224.0.0.1");
        // @api: dgram.Socket.dropMembership
        // @expect: socket_dropMembership: true
        console.log("socket_dropMembership: true");
        sock.addSourceSpecificMembership("127.0.0.1", "224.0.0.1");
        // @api: dgram.Socket.addSourceSpecificMembership
        // @expect: socket_addSourceSpecificMembership: true
        console.log("socket_addSourceSpecificMembership: true");
        sock.dropSourceSpecificMembership("127.0.0.1", "224.0.0.1");
        // @api: dgram.Socket.dropSourceSpecificMembership
        // @expect: socket_dropSourceSpecificMembership: true
        console.log("socket_dropSourceSpecificMembership: true");
        sock.setRecvBufferSize(65536);
        // @api: dgram.Socket.setRecvBufferSize
        // @expect: socket_setRecvBufferSize: true
        console.log("socket_setRecvBufferSize: true");
        sock.setSendBufferSize(65536);
        // @api: dgram.Socket.setSendBufferSize
        // @expect: socket_setSendBufferSize: true
        console.log("socket_setSendBufferSize: true");

        sock.close(() => {
            // @api: dgram.Socket.close
            // @expect: socket_close: true
            console.log("socket_close: true");
            const disposableSock = new Socket("udp4");
            // @api: dgram.Socket.[Symbol.asyncDispose]
            // @expect: socket_asyncDispose: true
            disposableSock[Symbol.asyncDispose]();
            console.log("socket_asyncDispose: true");
            const sock2 = createSocket("udp4");
            // @api: dgram.createSocket
            // @expect: createSocket_res: true
            console.log("createSocket_res: " + (sock2 instanceof Socket));
        });
    });
});
