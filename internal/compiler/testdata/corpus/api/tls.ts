import {
    DEFAULT_ECDH_CURVE,
    DEFAULT_MAX_VERSION,
    DEFAULT_MIN_VERSION,
    DEFAULT_CIPHERS,
    rootCertificates,
    checkServerIdentity,
    createSecureContext,
    getCACertificates,
    getCiphers,
    Server,
    createServer,
    TLSSocket
} from "node:tls";
import { readFileSync } from "node:fs";

const pem = readFileSync("testdata/corpus/api/tls-cert.pem", "utf8");
const key = readFileSync("testdata/corpus/api/tls-key.pem", "utf8");

// @api: tls.DEFAULT_ECDH_CURVE
// @expect: tls_ecdh: auto
console.log("tls_ecdh: " + DEFAULT_ECDH_CURVE);

// @api: tls.DEFAULT_MAX_VERSION
// @expect: tls_max_ver: TLSv1.3
console.log("tls_max_ver: " + DEFAULT_MAX_VERSION);

// @api: tls.DEFAULT_MIN_VERSION
// @expect: tls_min_ver: TLSv1.2
console.log("tls_min_ver: " + DEFAULT_MIN_VERSION);

// @api: tls.DEFAULT_CIPHERS
// @expect: tls_ciphers_def: true
console.log("tls_ciphers_def: " + (DEFAULT_CIPHERS.length > 0));

// @api: tls.rootCertificates
// @expect: tls_roots: true
console.log("tls_roots: " + Array.isArray(rootCertificates));

// @api: tls.getCACertificates("bundled")
// @expect: tls_bundled_roots: true
const bundledCertificates = getCACertificates("bundled");
console.log("tls_bundled_roots: " + (bundledCertificates.length > 0 && bundledCertificates.length === rootCertificates.length && bundledCertificates[0] === rootCertificates[0]));

// @api: tls.getCACertificates("system")
// @expect: tls_system_roots: true
const systemCertificates = getCACertificates("system");
console.log("tls_system_roots: " + (Array.isArray(systemCertificates) && systemCertificates.length > 0));

// @api: tls.checkServerIdentity
// @expect: tls_checkIdentity: true
console.log("tls_checkIdentity: " + (checkServerIdentity("unknown.example", { subject: "/CN=example.com" }) !== undefined));

// @api: tls.createSecureContext
// @expect: tls_secureContext: true
console.log("tls_secureContext: " + (createSecureContext() instanceof Object));
// @api: tls.getCACertificates
// @expect: tls_ca_certs: 1
console.log("tls_ca_certs: " + getCACertificates().length);

// @api: tls.getCiphers
// @expect: tls_ciphers_list: true
console.log("tls_ciphers_list: " + (getCiphers().length > 0));
// @expect: tls_ciphers_lowercase: true
const ciphers = getCiphers();
console.log("tls_ciphers_lowercase: " + (ciphers.length > 0 && ciphers[0] === ciphers[0].toLowerCase()));

const context = createSecureContext({ cert: pem, key: key });

// @api: tls.tls.Server
// @api: tls.createServer
// @expect: tls_server_inst: true
const srv = createServer({ cert: pem, key: key });
console.log("tls_server_inst: " + (srv instanceof Server));

// @api: tls.Server.addContext
// @expect: tls_srv_addContext: true
srv.addContext("example.com", context);
console.log("tls_srv_addContext: true");

// @api: tls.Server.getTicketKeys
// @expect: tls_srv_ticketKeys: 48
console.log("tls_srv_ticketKeys: " + srv.getTicketKeys().length);
// @expect: tls_srv_ticketKeys_type: true
console.log("tls_srv_ticketKeys_type: " + Buffer.isBuffer(srv.getTicketKeys()));
// @expect: tls_srv_ticketKeys_copy: true
const ticketKeys = srv.getTicketKeys();
const ticketKey = ticketKeys[0];
ticketKeys[0] = ticketKey ^ 255;
console.log("tls_srv_ticketKeys_copy: " + (srv.getTicketKeys()[0] === ticketKey));

// @api: tls.Server.setSecureContext
// @expect: tls_srv_setSecureContext: true
srv.setSecureContext({ cert: pem, key: key });
console.log("tls_srv_setSecureContext: true");

// @api: tls.Server.setTicketKeys
// @expect: tls_srv_setTicketKeys: true
srv.setTicketKeys(new Uint8Array(48));
console.log("tls_srv_setTicketKeys: true");

// @api: tls.tls.TLSSocket
// @api: tls.connect
// @expect: tls_socket_inst: true
const sock = new TLSSocket();
console.log("tls_socket_inst: " + (sock instanceof TLSSocket));

// @api: tls.TLSSocket.disableRenegotiation
// @expect: tls_sock_setServername: true
// @expect: tls_sock_disReneg: true
const optionSock = new TLSSocket();
optionSock.setServername("example.com");
console.log("tls_sock_setServername: true");
optionSock.disableRenegotiation();
console.log("tls_sock_disReneg: true");

// @api: tls.TLSSocket.enableTrace
// @expect: tls_sock_enTrace: true
optionSock.enableTrace();
console.log("tls_sock_enTrace: true");
optionSock.destroy();

// @api: tls.TLSSocket.authorizationError
// @expect: tls_sock_authErr: true
console.log("tls_sock_authErr: " + (sock.authorizationError === null));

// @api: tls.TLSSocket.authorized
// @expect: tls_sock_auth: false
console.log("tls_sock_auth: " + sock.authorized);

// @api: tls.TLSSocket.encrypted
// @expect: tls_sock_enc: true
console.log("tls_sock_enc: " + sock.encrypted);

// @api: tls.TLSSocket.localAddress
// @expect: tls_sock_locAddr: true
console.log("tls_sock_locAddr: " + (sock.localAddress === undefined));

// @api: tls.TLSSocket.localPort
// @expect: tls_sock_locPort: undefined
console.log("tls_sock_locPort: " + sock.localPort);

// @api: tls.TLSSocket.remoteAddress
// @expect: tls_sock_remAddr: true
console.log("tls_sock_remAddr: " + (sock.remoteAddress === undefined));

// @api: tls.TLSSocket.remoteFamily
// @expect: tls_sock_remFam: true
console.log("tls_sock_remFam: " + (sock.remoteFamily === undefined));

// @api: tls.TLSSocket.localFamily
// @expect: tls_sock_locFam: true
console.log("tls_sock_locFam: " + (sock.localFamily === undefined));

// @api: tls.TLSSocket.remotePort
// @expect: tls_sock_remPort: undefined
console.log("tls_sock_remPort: " + sock.remotePort);

// @api: tls.TLSSocket.address
// @expect: tls_sock_addr: undefined
console.log("tls_sock_addr: " + sock.address().port);
// @expect: tls_sock_addr_empty: true
console.log("tls_sock_addr_empty: " + (Object.keys(sock.address()).length === 0));

// @api: tls.TLSSocket.getCertificate
// @expect: tls_sock_getCert: true
console.log("tls_sock_getCert: " + (typeof sock.getCertificate() === "object"));
// @expect: tls_sock_getCert_empty: true
console.log("tls_sock_getCert_empty: " + (Object.keys(sock.getCertificate()).length === 0));

// @api: tls.TLSSocket.getCipher
// @expect: tls_sock_getCipher: false
console.log("tls_sock_getCipher: " + (typeof sock.getCipher() === "object"));

// @api: tls.TLSSocket.getEphemeralKeyInfo
// @expect: tls_sock_getEph: true
console.log("tls_sock_getEph: " + (typeof sock.getEphemeralKeyInfo() === "object"));
// @expect: tls_sock_getEph_empty: false
console.log("tls_sock_getEph_empty: " + (Object.keys(sock.getEphemeralKeyInfo()).length === 0));

// @api: tls.TLSSocket.getFinished
// @expect: tls_sock_getFin: undefined
const finished = sock.getFinished();
console.log("tls_sock_getFin: " + (finished === undefined || finished === null ? "undefined" : finished.length));

// @api: tls.TLSSocket.getPeerCertificate
// @expect: tls_sock_getPeerCert: true
console.log("tls_sock_getPeerCert: " + (typeof sock.getPeerCertificate() === "object"));
// @expect: tls_sock_getPeerCert_empty: true
console.log("tls_sock_getPeerCert_empty: " + (Object.keys(sock.getPeerCertificate()).length === 0));

// @api: tls.TLSSocket.getPeerFinished
// @expect: tls_sock_getPeerFin: undefined
const peerFinished = sock.getPeerFinished();
console.log("tls_sock_getPeerFin: " + (peerFinished === undefined || peerFinished === null ? "undefined" : peerFinished.length));

// @api: tls.TLSSocket.getPeerX509Certificate
// @expect: tls_sock_getPeerX509: true
console.log("tls_sock_getPeerX509: " + (sock.getPeerX509Certificate() === undefined));

// @api: tls.TLSSocket.getProtocol
// @expect: tls_sock_getProto: true
console.log("tls_sock_getProto: " + (sock.getProtocol() === DEFAULT_MAX_VERSION));

// @api: tls.TLSSocket.getSession
// @expect: tls_sock_getSess: undefined
const session = sock.getSession();
console.log("tls_sock_getSess: " + (session === undefined || session === null ? "undefined" : session.length));

// @api: tls.TLSSocket.getSharedSigalgs
// @expect: tls_sock_getSigalgs: 0
const sharedSigalgs = sock.getSharedSigalgs();
console.log("tls_sock_getSigalgs: " + (sharedSigalgs === null ? 0 : sharedSigalgs.length));

// @api: tls.TLSSocket.getTLSTicket
// @expect: tls_sock_getTicket: undefined
const ticket = sock.getTLSTicket();
console.log("tls_sock_getTicket: " + (ticket === undefined || ticket === null ? "undefined" : ticket.length));

// @api: tls.TLSSocket.getX509Certificate
// @expect: tls_sock_getX509: true
console.log("tls_sock_getX509: " + (sock.getX509Certificate() === undefined));

// @api: tls.TLSSocket.isSessionReused
// @expect: tls_sock_isReused: false
console.log("tls_sock_isReused: " + sock.isSessionReused());

// @api: tls.TLSSocket.setKeyCert
// @expect: tls_sock_setKeyCert: true
sock.setKeyCert({ cert: pem, key: key });
console.log("tls_sock_setKeyCert: true");

// @api: tls.TLSSocket.renegotiate
// @expect: tls_sock_reneg: true
console.log("tls_sock_reneg: " + sock.renegotiate({}, () => {}));

// @api: tls.TLSSocket.setMaxSendFragment
// @expect: tls_sock_setMaxSendFrag: true
console.log("tls_sock_setMaxSendFrag: " + sock.setMaxSendFragment(16384));
sock.destroy();
