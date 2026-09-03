import { createServer as createNetServer } from "node:net";
import { createServer as createTLSServer } from "node:tls";

const netServer = createNetServer();
const tlsServer = createTLSServer();

// @expect: tls_function_identity: true
console.log("tls_function_identity: " + (netServer instanceof Object && tlsServer instanceof Object));
