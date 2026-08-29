// Node.js TLS module (node:tls)

export interface AddressInfo {
    port: number;
    family: string;
    address: string;
}

export interface CipherInfo {
    name: string;
    version: string;
}

export interface EphemeralKeyInfo {
    type: string;
    name: string;
    size: number;
}

export const DEFAULT_ECDH_CURVE = "auto";
export const DEFAULT_MAX_VERSION = "TLSv1.3";
export const DEFAULT_MIN_VERSION = "TLSv1.2";
export const DEFAULT_CIPHERS = "TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256:TLS_AES_128_GCM_SHA256";
export const rootCertificates: string[] = [];

let _caCertificates: string[] = [];

export function checkServerIdentity(hostname: string, cert: unknown): unknown {
    return undefined;
}

export function createSecureContext(options?: unknown): Record<string, unknown> {
    return {};
}

export function setDefaultCACertificates(certs: string[]): void {
    _caCertificates = certs;
}

export function getCACertificates(): string[] {
    return _caCertificates;
}

export function getCiphers(): string[] {
    return [
        "TLS_AES_256_GCM_SHA384",
        "TLS_CHACHA20_POLY1305_SHA256",
        "TLS_AES_128_GCM_SHA256"
    ];
}

export class SecurePair {
    cleartext: unknown = null;
    encrypted: unknown = null;

    constructor() {
        this.cleartext = null;
        this.encrypted = null;
    }
}

export function createSecurePair(context?: unknown, isServer?: boolean, requestCert?: boolean, rejectUnauthorized?: boolean): SecurePair {
    return new SecurePair();
}

export class Server {
    constructor(options?: unknown, listener?: unknown) {}

    addContext(hostname: string, context: unknown): void {}

    address(): AddressInfo {
        return { port: 443, family: "IPv4", address: "127.0.0.1" };
    }

    close(callback?: unknown): Server {
        if (typeof callback === "function") {
            (callback as Function)();
        }
        return this;
    }

    getTicketKeys(): Uint8Array {
        return new Uint8Array(48);
    }

    listen(port?: unknown, host?: unknown, callback?: unknown): Server {
        const cb = typeof port === "function" ? port : (typeof host === "function" ? host : callback);
        if (typeof cb === "function") {
            (cb as Function)();
        }
        return this;
    }

    setSecureContext(options: unknown): void {}

    setTicketKeys(keys: unknown): void {}
}

export function createServer(options?: unknown, secureConnectionListener?: unknown): Server {
    return new Server(options, secureConnectionListener);
}

export class TLSSocket {
    authorizationError: unknown = null;
    authorized: boolean = true;
    encrypted: boolean = true;
    localAddress: string = "127.0.0.1";
    localPort: number = 0;
    remoteAddress: string = "127.0.0.1";
    remoteFamily: string = "IPv4";
    remotePort: number = 443;

    constructor(socket?: unknown, options?: unknown) {}

    address(): AddressInfo {
        return { port: this.localPort, family: "IPv4", address: this.localAddress };
    }

    disableRenegotiation(): void {}

    enableTrace(): void {}

    exportKeyingMaterial(length: number, label: string, context?: unknown): Uint8Array {
        return new Uint8Array(length);
    }

    getCertificate(): Record<string, unknown> {
        return {};
    }

    getCipher(): CipherInfo {
        return { name: "TLS_AES_256_GCM_SHA384", version: "TLSv1.3" };
    }

    getEphemeralKeyInfo(): EphemeralKeyInfo {
        return { type: "ECDH", name: "prime256v1", size: 256 };
    }

    getFinished(): Uint8Array {
        return new Uint8Array(0);
    }

    getPeerCertificate(detailed?: boolean): Record<string, unknown> {
        return { subject: {} };
    }

    getPeerFinished(): Uint8Array {
        return new Uint8Array(0);
    }

    getPeerX509Certificate(): unknown {
        return null;
    }

    getProtocol(): string {
        return "TLSv1.3";
    }

    getSession(): Uint8Array {
        return new Uint8Array(0);
    }

    getSharedSigalgs(): string[] {
        return ["ecdsa_secp256r1_sha256", "rsa_pss_rsae_sha256"];
    }

    getTLSTicket(): Uint8Array {
        return new Uint8Array(0);
    }

    getX509Certificate(): unknown {
        return null;
    }

    isSessionReused(): boolean {
        return false;
    }

    renegotiate(options: unknown, callback?: unknown): boolean {
        if (typeof callback === "function") {
            (callback as Function)();
        }
        return true;
    }

    setKeyCert(options: unknown): void {}

    setMaxSendFragment(size: number): boolean {
        return true;
    }
}

export function connect(options?: unknown, callback?: unknown): TLSSocket {
    const sock = new TLSSocket(null, options);
    if (typeof callback === "function") {
        (callback as Function)();
    }
    return sock;
}

export default {
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
    connect,
};
