import { createSecureContext, createSecurePair } from "node:tls";
import { readFileSync } from "node:fs";

const pem = readFileSync("testdata/corpus/api/tls-cert.pem", "utf8");
const key = readFileSync("testdata/corpus/api/tls-key.pem", "utf8");

// @run.err: TLS pair handshake failed
const serverContext = createSecureContext({ cert: pem, key: key });
const clientContext = createSecureContext({ ca: [] });
const client = createSecurePair(clientContext, false, false, true);
const server = createSecurePair(serverContext, true, false, false);

for (let i = 0; i < 4; i++) {
    const clientHello = client.encrypted.read();
    if (clientHello.length > 0) server.encrypted.write(clientHello);
    const serverHello = server.encrypted.read();
    if (serverHello.length > 0) client.encrypted.write(serverHello);
}
