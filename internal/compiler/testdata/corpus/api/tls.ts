import {
    DEFAULT_ECDH_CURVE,
    DEFAULT_MAX_VERSION,
    DEFAULT_MIN_VERSION,
    DEFAULT_CIPHERS,
    rootCertificates,
    checkServerIdentity,
    createSecureContext,
    setDefaultCACertificates,
    getCACertificates,
    getCiphers,
    SecurePair,
    createSecurePair,
    Server,
    createServer,
    TLSSocket,
    connect
} from "node:tls";

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

// @api: tls.checkServerIdentity
// @expect: tls_checkIdentity: undefined
console.log("tls_checkIdentity: " + checkServerIdentity("localhost", {}));

// @api: tls.createSecureContext
// @expect: tls_secureContext: true
console.log("tls_secureContext: " + (typeof createSecureContext() === "object"));

// @api: tls.setDefaultCACertificates
// @api: tls.getCACertificates
// @expect: tls_ca_certs: 1
setDefaultCACertificates(["cert1"]);
console.log("tls_ca_certs: " + getCACertificates().length);

// @api: tls.getCiphers
// @expect: tls_ciphers_list: true
console.log("tls_ciphers_list: " + (getCiphers().length > 0));

// @api: tls.tls.SecurePair
// @api: tls.createSecurePair
// @expect: tls_secPair: true
const sp = createSecurePair();
console.log("tls_secPair: " + (sp instanceof SecurePair));

// @api: tls.tls.Server
// @api: tls.createServer
// @expect: tls_server_inst: true
const srv = createServer();
console.log("tls_server_inst: " + (srv instanceof Server));

// @api: tls.Server.addContext
// @expect: tls_srv_addContext: true
srv.addContext("example.com", {});
console.log("tls_srv_addContext: true");

// @api: tls.Server.address
// @expect: tls_srv_address: 443
console.log("tls_srv_address: " + srv.address().port);

// @api: tls.Server.getTicketKeys
// @expect: tls_srv_ticketKeys: 48
console.log("tls_srv_ticketKeys: " + srv.getTicketKeys().length);

// @api: tls.Server.setSecureContext
// @expect: tls_srv_setSecureContext: true
srv.setSecureContext({});
console.log("tls_srv_setSecureContext: true");

// @api: tls.Server.setTicketKeys
// @expect: tls_srv_setTicketKeys: true
srv.setTicketKeys(new Uint8Array(48));
console.log("tls_srv_setTicketKeys: true");

// @api: tls.Server.listen
// @api: tls.Server.close
// @expect: tls_srv_listen_close: true
srv.listen(() => {});
srv.close();
console.log("tls_srv_listen_close: true");

// @api: tls.tls.TLSSocket
// @api: tls.connect
// @expect: tls_socket_inst: true
const sock = connect({ port: 443 });
console.log("tls_socket_inst: " + (sock instanceof TLSSocket));

// @api: tls.TLSSocket.authorizationError
// @expect: tls_sock_authErr: true
console.log("tls_sock_authErr: " + (sock.authorizationError === null));

// @api: tls.TLSSocket.authorized
// @expect: tls_sock_auth: true
console.log("tls_sock_auth: " + sock.authorized);

// @api: tls.TLSSocket.encrypted
// @expect: tls_sock_enc: true
console.log("tls_sock_enc: " + sock.encrypted);

// @api: tls.TLSSocket.localAddress
// @expect: tls_sock_locAddr: 127.0.0.1
console.log("tls_sock_locAddr: " + sock.localAddress);

// @api: tls.TLSSocket.localPort
// @expect: tls_sock_locPort: 0
console.log("tls_sock_locPort: " + sock.localPort);

// @api: tls.TLSSocket.remoteAddress
// @expect: tls_sock_remAddr: 127.0.0.1
console.log("tls_sock_remAddr: " + sock.remoteAddress);

// @api: tls.TLSSocket.remoteFamily
// @expect: tls_sock_remFam: IPv4
console.log("tls_sock_remFam: " + sock.remoteFamily);

// @api: tls.TLSSocket.remotePort
// @expect: tls_sock_remPort: 443
console.log("tls_sock_remPort: " + sock.remotePort);

// @api: tls.TLSSocket.address
// @expect: tls_sock_addr: 0
console.log("tls_sock_addr: " + sock.address().port);

// @api: tls.TLSSocket.disableRenegotiation
// @expect: tls_sock_disReneg: true
sock.disableRenegotiation();
console.log("tls_sock_disReneg: true");

// @api: tls.TLSSocket.enableTrace
// @expect: tls_sock_enTrace: true
sock.enableTrace();
console.log("tls_sock_enTrace: true");

// @api: tls.TLSSocket.exportKeyingMaterial
// @expect: tls_sock_expKey: 32
console.log("tls_sock_expKey: " + sock.exportKeyingMaterial(32, "test").length);

// @api: tls.TLSSocket.getCertificate
// @expect: tls_sock_getCert: true
console.log("tls_sock_getCert: " + (typeof sock.getCertificate() === "object"));

// @api: tls.TLSSocket.getCipher
// @expect: tls_sock_getCipher: TLSv1.3
console.log("tls_sock_getCipher: " + sock.getCipher().version);

// @api: tls.TLSSocket.getEphemeralKeyInfo
// @expect: tls_sock_getEph: ECDH
console.log("tls_sock_getEph: " + sock.getEphemeralKeyInfo().type);

// @api: tls.TLSSocket.getFinished
// @expect: tls_sock_getFin: 0
console.log("tls_sock_getFin: " + sock.getFinished().length);

// @api: tls.TLSSocket.getPeerCertificate
// @expect: tls_sock_getPeerCert: true
console.log("tls_sock_getPeerCert: " + (typeof sock.getPeerCertificate() === "object"));

// @api: tls.TLSSocket.getPeerFinished
// @expect: tls_sock_getPeerFin: 0
console.log("tls_sock_getPeerFin: " + sock.getPeerFinished().length);

// @api: tls.TLSSocket.getPeerX509Certificate
// @expect: tls_sock_getPeerX509: true
console.log("tls_sock_getPeerX509: " + (sock.getPeerX509Certificate() === null));

// @api: tls.TLSSocket.getProtocol
// @expect: tls_sock_getProto: TLSv1.3
console.log("tls_sock_getProto: " + sock.getProtocol());

// @api: tls.TLSSocket.getSession
// @expect: tls_sock_getSess: 0
console.log("tls_sock_getSess: " + sock.getSession().length);

// @api: tls.TLSSocket.getSharedSigalgs
// @expect: tls_sock_getSigalgs: true
console.log("tls_sock_getSigalgs: " + (sock.getSharedSigalgs().length > 0));

// @api: tls.TLSSocket.getTLSTicket
// @expect: tls_sock_getTicket: 0
console.log("tls_sock_getTicket: " + sock.getTLSTicket().length);

// @api: tls.TLSSocket.getX509Certificate
// @expect: tls_sock_getX509: true
console.log("tls_sock_getX509: " + (sock.getX509Certificate() === null));

// @api: tls.TLSSocket.isSessionReused
// @expect: tls_sock_isReused: false
console.log("tls_sock_isReused: " + sock.isSessionReused());

// @api: tls.TLSSocket.renegotiate
// @expect: tls_sock_reneg: true
console.log("tls_sock_reneg: " + sock.renegotiate({}));

// @api: tls.TLSSocket.setKeyCert
// @expect: tls_sock_setKeyCert: true
sock.setKeyCert({});
console.log("tls_sock_setKeyCert: true");

// @api: tls.TLSSocket.setMaxSendFragment
// @expect: tls_sock_setMaxSendFrag: true
console.log("tls_sock_setMaxSendFrag: " + sock.setMaxSendFragment(16384));
