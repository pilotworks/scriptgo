import { SecurePair, createSecurePair, createSecureContext, X509Certificate, TLSSocket, createServer } from "node:tls";
import { readFileSync } from "node:fs";

const pem = readFileSync("testdata/corpus/api/tls-cert.pem", "utf8");
const key = readFileSync("testdata/corpus/api/tls-key.pem", "utf8");

// @api: tls.X509Certificate
// @native.expected: tls_x509_cert_fingerprint: true
const cert = new X509Certificate(pem);
console.log("tls_x509_cert_fingerprint: " + (cert.fingerprint256.length > 0));

// @api: tls.X509Certificate(DataView)
// @native.expected: tls_x509_dataView: true
const certView = new DataView(cert.raw.buffer, cert.raw.byteOffset, cert.raw.byteLength);
const certFromView = new X509Certificate(certView);
const certFromBuffer = new X509Certificate(cert.raw);
console.log("tls_x509_dataView: " + (certFromView.fingerprint256 === cert.fingerprint256 && certFromBuffer.fingerprint256 === cert.fingerprint256));

// @api: tls.X509Certificate.checkHost
// @native.expected: tls_x509_checkHost: true
console.log("tls_x509_checkHost: " + (cert.checkHost("unknown.com") === undefined));

// @native.expected: tls_x509_checkHost_case: true
console.log("tls_x509_checkHost_case: " + (cert.checkHost("EXAMPLE.COM") === "example.com"));

// @native.expected: tls_x509_checkHost_options: true
console.log("tls_x509_checkHost_options: " + (cert.checkHost("EXAMPLE.COM", { wildcards: false, subject: "never" }) === "example.com"));

// @native.expected: tls_x509_checkEmail_options: true
console.log("tls_x509_checkEmail_options: " + (cert.checkEmail("test@example.com", { subject: "never" }) === undefined));

// @native.expected: tls_x509_checkIP_no_options: true
console.log("tls_x509_checkIP_no_options: " + (cert.checkIP("127.0.0.1") === undefined));

// @api: tls.X509Certificate.toJSON
// @native.expected: tls_x509_toJSON: true
console.log("tls_x509_toJSON: " + (cert.toJSON().length > 0));

// @native.expected: tls_x509_toJSON_pem: true
console.log("tls_x509_toJSON_pem: " + cert.toJSON().startsWith("-----BEGIN CERTIFICATE-----"));

const context = createSecureContext({ cert: pem, key: key });

// @api: tls.tls.SecurePair
// @api: tls.createSecurePair
// @native.expected: tls_secPair: true
const sp = createSecurePair(context);
console.log("tls_secPair: " + (sp instanceof SecurePair));
sp.cleartext.destroy();
sp.encrypted.destroy();

// @api: tls.SecurePair handshake and cleartext round-trip
// @native.expected: tls_pair_roundtrip: true
const clientPair = createSecurePair(context, false, false, false);
const serverPair = createSecurePair(context, true, false, false);
for (let i = 0; i < 4; i++) {
    const clientHello = clientPair.encrypted.read();
    if (clientHello.length > 0) serverPair.encrypted.write(clientHello);
    const serverHello = serverPair.encrypted.read();
    if (serverHello.length > 0) clientPair.encrypted.write(serverHello);
}
clientPair.cleartext.write("hello");
const applicationData = clientPair.encrypted.read();
if (applicationData.length > 0) serverPair.encrypted.write(applicationData);
console.log("tls_pair_roundtrip: " + (serverPair.cleartext.read() === "hello"));

// @native.expected: tls_pair_empty_write: true
const emptyStringWrite = clientPair.cleartext.write("");
const emptyBytesWrite = clientPair.cleartext.write(new Uint8Array(0));
console.log("tls_pair_empty_write: " + (emptyStringWrite && emptyBytesWrite));

// @native.expected: tls_pair_utf8_roundtrip: true
clientPair.cleartext.write("hé");
const unicodeApplicationData = clientPair.encrypted.read();
if (unicodeApplicationData.length > 0) serverPair.encrypted.write(unicodeApplicationData);
console.log("tls_pair_utf8_roundtrip: " + (serverPair.cleartext.read() === "hé"));

// @native.expected: tls_pair_cipher: true
const clientCipher = clientPair.cleartext.getCipher();
const serverCipher = serverPair.cleartext.getCipher();
console.log("tls_pair_cipher: " + (clientCipher !== undefined && clientCipher !== null && serverCipher !== undefined && serverCipher !== null && clientCipher.name === serverCipher.name));

// @api: tls.TLSSocket.getSession TLS 1.3
// @native.expected: tls_pair_session: true
const tlsSession = clientPair.cleartext.getSession();
console.log("tls_pair_session: " + (tlsSession !== undefined && tlsSession !== null && tlsSession.length > 0));

// @native.expected: tls_pair_sigalgs: true
const pairSigalgs = serverPair.cleartext.getSharedSigalgs();
console.log("tls_pair_sigalgs: " + (pairSigalgs !== null && pairSigalgs.length > 0));

// @api: tls.TLSSocket.exportKeyingMaterial
// @native.expected: tls_pair_expKey: 32
console.log("tls_pair_expKey: " + clientPair.cleartext.exportKeyingMaterial(32, "test", new Uint8Array(0)).length);

// @native.expected: tls_pair_expKey_context: true
const exportContext = Buffer.from("context");
const clientExported = clientPair.cleartext.exportKeyingMaterial(32, "test", exportContext);
const serverExported = serverPair.cleartext.exportKeyingMaterial(32, "test", exportContext);
let exportContextEqual = clientExported.length === 32 && serverExported.length === 32;
for (let i = 0; i < clientExported.length && exportContextEqual; i++) {
    if (clientExported[i] !== serverExported[i]) exportContextEqual = false;
}
console.log("tls_pair_expKey_context: " + exportContextEqual);

// @api: tls.TLSSocket.setSession
// @native.expected: tls_pair_setSession: true
const sessionContext = createSecureContext({ cert: pem, key: key, minVersion: "TLSv1.2", maxVersion: "TLSv1.2" });
const sessionClientPair = createSecurePair(sessionContext, false, false, false);
const sessionServerPair = createSecurePair(sessionContext, true, false, false);
for (let i = 0; i < 4; i++) {
    const sessionClientHello = sessionClientPair.encrypted.read();
    if (sessionClientHello.length > 0) sessionServerPair.encrypted.write(sessionClientHello);
    const sessionServerHello = sessionServerPair.encrypted.read();
    if (sessionServerHello.length > 0) sessionClientPair.encrypted.write(sessionServerHello);
}
sessionClientPair.cleartext.write("session");
const sessionApplicationData = sessionClientPair.encrypted.read();
if (sessionApplicationData.length > 0) sessionServerPair.encrypted.write(sessionApplicationData);
sessionServerPair.cleartext.read();
const pairSession = sessionClientPair.cleartext.getSession();
const sessionSocket = new TLSSocket(null, { secureContext: sessionContext });
let sessionSet = false;
if (pairSession !== undefined && pairSession !== null && pairSession.length > 0) {
    sessionSocket.setSession(pairSession);
    sessionSet = true;
}
console.log("tls_pair_setSession: " + sessionSet);
sessionSocket.destroy();
sessionClientPair.cleartext.destroy();
sessionClientPair.encrypted.destroy();
sessionServerPair.cleartext.destroy();
sessionServerPair.encrypted.destroy();

// @api: tls.TLSSocket.getPeerCertificate(detailed)
// @native.expected: tls_peer_cert_detail: true
// @native.expected: tls_peer_cert_fields: true
const peerCertificate = clientPair.cleartext.getPeerCertificate();
const detailedPeerCertificate = clientPair.cleartext.getPeerCertificate(true);
console.log("tls_peer_cert_detail: " + (peerCertificate !== null && detailedPeerCertificate !== null && peerCertificate.pem === undefined && typeof detailedPeerCertificate.pem === "string"));
if (peerCertificate !== null) {
    const peerSubject = peerCertificate.subject as { CN?: unknown };
    console.log("tls_peer_cert_fields: " + (peerSubject.CN === "example.com" && peerCertificate.subjectaltname === "DNS:example.com" && Buffer.isBuffer(peerCertificate.raw) && typeof peerCertificate.fingerprint256 === "string" && (peerCertificate.fingerprint256 as string).indexOf(":") >= 0));
}
clientPair.cleartext.destroy();
clientPair.encrypted.destroy();
serverPair.cleartext.destroy();
serverPair.encrypted.destroy();

// @api: tls.Server.listen
// @api: tls.Server.address
// @api: tls.Server.close
// @native.expected: tls_srv_listen_close: true
const server = createServer({ cert: pem, key: key });
server.listen(0, "127.0.0.1");
const address = server.address();
console.log("tls_srv_listen_close: " + (address !== null && address.port > 0));
server.close();
