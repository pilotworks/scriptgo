import { Socket } from "node:net";
import { TLSSocket } from "node:tls";

const socket = new Socket();
const tls = new TLSSocket(socket);

// @expect: tls_net_socket: true
console.log("tls_net_socket: " + tls.encrypted);
