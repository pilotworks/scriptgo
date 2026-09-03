// Node.js TLS module (node:tls)

import type { Socket } from "node:net";
import type { Duplex } from "node:stream";

export interface AddressInfo {
    port: number;
    family: string;
    address: string;
}

interface SocketAddressInfo {
    port?: number;
    family?: string;
    address?: string;
}

export interface CipherInfo {
    name: string;
    version: string;
}

type CertificateName = Record<string, string | string[]>;

export interface PeerCertificate {
    subject?: string | CertificateName;
    issuer?: string | CertificateName;
    subjectaltname?: string;
    valid_from?: string;
    valid_to?: string;
    fingerprint?: string;
    fingerprint256?: string;
    fingerprint512?: string;
    serialNumber?: string;
    ca?: boolean;
    pem?: string;
    raw?: Buffer;
}

export class EphemeralKeyInfo {
    type: string | undefined = undefined;
    name: string | undefined = undefined;
    size: number | undefined = undefined;
}

export type TLSVersion = "TLSv1.3" | "TLSv1.2" | "TLSv1.1" | "TLSv1";
export type TLSBinary = string | ArrayBufferView;
type TLSCertificateList = TLSBinary | ReadonlyArray<TLSBinary>;
type X509SubjectOption = "always" | "default" | "never";

export interface SecureContextOptions {
    ca?: TLSCertificateList;
    cert?: TLSBinary;
    key?: TLSBinary;
    minVersion?: TLSVersion;
    maxVersion?: TLSVersion;
    ciphers?: string;
}

export interface X509CheckOptions {
    subject?: X509SubjectOption;
    wildcards?: boolean;
    partialWildcards?: boolean;
    multiLabelWildcards?: boolean;
    singleLabelSubdomains?: boolean;
}

export interface TLSSocketOptions extends SecureContextOptions {
    rejectUnauthorized?: boolean;
    requestCert?: boolean;
    secureContext?: SecureContext;
    session?: Uint8Array;
    isServer?: boolean;
}

export interface TLSConnectionOptions extends SecureContextOptions {
    host?: string;
    port?: number;
    rejectUnauthorized?: boolean;
    secureContext?: SecureContext;
    servername?: string;
    session?: Uint8Array;
}

export interface TLSServerOptions extends SecureContextOptions {
    rejectUnauthorized?: boolean;
    requestCert?: boolean;
}

export interface TLSListenOptions {
    port?: number;
    host?: string;
    backlog?: number;
}

interface InternalTLSSocketOptions extends TLSSocketOptions {
    servername?: string;
    __tlsHandle?: number;
    __tlsPair?: boolean;
    __tlsMode?: number;
    __tlsConnected?: boolean;
}

type TLSOptionRecord = {
    cert?: TLSBinary;
    key?: TLSBinary;
    ca?: TLSCertificateList;
    minVersion?: TLSVersion;
    maxVersion?: TLSVersion;
    ciphers?: string;
    host?: string;
    port?: number;
    rejectUnauthorized?: boolean;
    requestCert?: boolean;
    isServer?: boolean;
    servername?: string;
    session?: Uint8Array;
    secureContext?: SecureContext;
};

type TLSCallback = () => void;
type TLSRenegotiationCallback = (err: Error | null) => void;
type TLSSecureConnectionListener = (socket: TLSSocket) => void;

class TLSEventListenerEntry {
    fn: Function;
    once: boolean;

    constructor(fn: Function, once: boolean) {
        this.fn = fn;
        this.once = once;
    }
}

class TLSEventBucket {
    name: string;
    listeners: TLSEventListenerEntry[];

    constructor(name: string, listeners: TLSEventListenerEntry[]) {
        this.name = name;
        this.listeners = listeners;
    }
}

type NativeCertificate = {
    subject?: string;
    issuer?: string;
    subjectAltName?: string;
    validFrom?: string;
    validTo?: string;
    fingerprint?: string;
    fingerprint256?: string;
    fingerprint512?: string;
    serialNumber?: string;
    pem?: string;
};

type NativeEphemeral = {
    type?: string;
    name?: string;
    size?: number;
    server?: boolean;
};

declare namespace __scriptgo {
    function randomBytes(size: number): Uint8Array;
    function tlsContextCreate(cert: string, key: string, ca: string, minVersion: string, maxVersion: string, ciphers: string, caProvided: boolean): number;
    function tlsSocketCreate(context: number, isServer: boolean): number;
    function tlsSocketConnect(context: number, host: string, port: number, servername: string, rejectUnauthorized: boolean, session: Uint8Array): number;
    function tlsSocketAdopt(context: number, fd: number, servername: string, isServer: boolean, requestCert: boolean, rejectUnauthorized: boolean, session: Uint8Array): number;
    function tlsSocketWrite(handle: number, data: string, length: number): number;
    function tlsSocketWriteBytes(handle: number, data: Uint8Array): number;
    function tlsSocketRead(handle: number, maxLength: number): string;
    function tlsPairWrite(handle: number, mode: number, data: string, length: number): number;
    function tlsPairWriteBytes(handle: number, mode: number, data: Uint8Array): number;
    function tlsPairRead(handle: number, mode: number, maxLength: number): string;
    function tlsSocketClose(handle: number): void;
    function tlsSocketInfo(handle: number, property: string): string;
    function tlsSocketNumber(handle: number, property: string): number;
    function tlsSocketBool(handle: number, property: string): boolean;
    function tlsExportKeyingMaterial(handle: number, length: number, label: string, context: Uint8Array | null): string;
    function tlsSocketSetOption(handle: number, option: string, value: number): void;
    function tlsSocketSetServername(handle: number, servername: string): void;
    function tlsSocketSetSession(handle: number, session: Uint8Array): void;
    function tlsSocketSetKeyCert(handle: number, cert: string, key: string): void;
    function tlsSocketRenegotiate(handle: number): boolean;
    function tlsPairCreate(context: number, isServer: boolean, requestCert: boolean, rejectUnauthorized: boolean): number;
    function tlsServerListen(context: number, requestCert: boolean, rejectUnauthorized: boolean, host: string, port: number, backlog: number): number;
    function tlsServerAccept(handle: number): number;
    function tlsServerClose(handle: number): void;
    function tlsServerInfo(handle: number, property: string): string;
    function tlsServerSetContext(handle: number, context: number, requestCert: boolean, rejectUnauthorized: boolean): void;
    function tlsServerAddContext(handle: number, hostname: string, context: number): void;
    function tlsServerSetTicketKeys(handle: number, hex: string): void;
    function tlsX509ParsePem(pem: string): string;
    function tlsX509ParseBytes(data: Uint8Array): string;
    function tlsCiphers(): string;
    function tlsRootCertificates(): string;
    function tlsSystemCertificates(): string;
    function tlsExtraCertificates(): string;
}

export const DEFAULT_ECDH_CURVE = "auto";
export const DEFAULT_MAX_VERSION = "TLSv1.3";
export const DEFAULT_MIN_VERSION = "TLSv1.2";
export const DEFAULT_CIPHERS = "TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256:TLS_AES_128_GCM_SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA384:DHE-RSA-AES256-SHA384:ECDHE-RSA-AES256-SHA256:DHE-RSA-AES256-SHA256:HIGH:!aNULL:!eNULL:!EXPORT:!DES:!RC4:!MD5:!PSK:!SRP:!CAMELLIA";
export const CLIENT_RENEG_LIMIT = 3;
export const CLIENT_RENEG_WINDOW = 600;
export const rootCertificates: string[] = JSON.parse(__scriptgo.tlsRootCertificates()) as string[];

let _caCertificates: string[] = [];
let _caCertificatesSet: boolean = false;

type TLSOptionValue = SecureContextOptions | TLSSocketOptions | TLSConnectionOptions | TLSServerOptions | TLSListenOptions | InternalTLSSocketOptions | TLSOptionRecord | null | undefined;

function optionRecord(value: TLSOptionValue): TLSOptionRecord {
    if (value !== null && typeof value === "object") {
        return value as TLSOptionRecord;
    }
    return {};
}

function optionString(value: string | undefined, fallback: string, name: string): string {
    if (value === undefined) return fallback;
    if (typeof value !== "string") throw new TypeError("TLS " + name + " must be a string");
    return value;
}

function byteView(value: ArrayBufferView, name: string, allowEmpty: boolean = true): Uint8Array {
    if (!ArrayBuffer.isView(value)) throw new TypeError("TLS " + name + " must be a TypedArray or DataView");
    const view = value as ArrayBufferView;
    if (!allowEmpty && view.byteLength === 0) throw new TypeError("TLS " + name + " must not be empty");
    return new Uint8Array(view.buffer, view.byteOffset, view.byteLength).slice();
}

function pemString(value: TLSBinary | null | undefined, name: string): string {
    if (value === undefined || value === null) return "";
    if (typeof value === "string") return value;
    if (ArrayBuffer.isView(value)) return Buffer.from(byteView(value, name)).toString("utf8");
    throw new TypeError("TLS " + name + " must be a string, Buffer, TypedArray, or DataView");
}

function isTLSBinary(value: TLSCertificateList | null | undefined): value is TLSBinary {
    return typeof value === "string" || ArrayBuffer.isView(value);
}

function certificateList(value: TLSCertificateList | null | undefined, name: string): string[] {
    if (value === null || value === undefined || isTLSBinary(value)) return [pemString(value, name)];
    const result: string[] = [];
    for (let i = 0; i < value.length; i++) {
        const certificate = pemString(value[i], name);
        let duplicate = false;
        for (let j = 0; j < result.length; j++) {
            if (result[j] === certificate) {
                duplicate = true;
                break;
            }
        }
        if (!duplicate) result.push(certificate);
    }
    return result;
}

function optionCA(options: TLSOptionRecord): string {
    if (options.ca !== undefined) return certificateList(options.ca, "ca").join("\n");
    return getCACertificates().join("\n");
}

function optionCAProvided(options: TLSOptionRecord): boolean {
    return options.ca !== undefined || _caCertificatesSet;
}

function bytesToHex(value: Uint8Array): string {
    let result = "";
    for (let i = 0; i < value.length; i++) {
        const byte = value[i] & 255;
        const hex = byte.toString(16);
        result += hex.length === 1 ? "0" + hex : hex;
    }
    return result;
}

function hexToBytes(value: string): Uint8Array {
    if (value.length === 0) return new Uint8Array(0);
    if (value.length % 2 !== 0) throw new Error("Native TLS returned an invalid hexadecimal value");
    const result = new Uint8Array(value.length / 2);
    for (let i = 0; i < result.length; i++) {
        const pair = value.substring(i * 2, i * 2 + 2);
        if (!/^[0-9a-fA-F]{2}$/.test(pair)) throw new Error("Native TLS returned an invalid hexadecimal value");
        result[i] = parseInt(pair, 16);
    }
    return result;
}

function pemToBytes(pem: string): Buffer {
    const begin = pem.indexOf("-----BEGIN CERTIFICATE-----");
    const end = pem.indexOf("-----END CERTIFICATE-----");
    if (begin < 0 || end <= begin) throw new Error("Native TLS returned an invalid PEM certificate");
    return Buffer.from(pem.substring(begin + 27, end), "base64");
}

function nativeObject(value: string): Record<string, unknown> {
    return JSON.parse(value) as Record<string, unknown>;
}

function nativeTLSCertificate(value: string): PeerCertificate {
    const parsed = nativeObject(value);
    if (Object.keys(parsed).length === 0) return {};
    const result: PeerCertificate = {
        subject: undefined,
        issuer: undefined,
        subjectaltname: undefined,
        valid_from: undefined,
        valid_to: undefined,
        fingerprint: undefined,
        fingerprint256: undefined,
        fingerprint512: undefined,
        serialNumber: undefined,
        ca: undefined,
        pem: undefined,
        raw: undefined,
    };
    const subjectObject = parsed.subjectObject as CertificateName | undefined;
    const issuerObject = parsed.issuerObject as CertificateName | undefined;
    const subject = parsed.subject as string | undefined;
    const issuer = parsed.issuer as string | undefined;
    const subjectAltName = parsed.subjectAltName as string | undefined;
    const validFrom = parsed.validFrom as string | undefined;
    const validTo = parsed.validTo as string | undefined;
    const fingerprint = parsed.fingerprint as string | undefined;
    const fingerprint256 = parsed.fingerprint256 as string | undefined;
    const fingerprint512 = parsed.fingerprint512 as string | undefined;
    const serialNumber = parsed.serialNumber as string | undefined;
    const ca = parsed.ca as boolean | undefined;
    const pem = parsed.pem as string | undefined;
    const raw = parsed.raw as string | undefined;
    if (subjectObject !== undefined) result.subject = subjectObject;
    else if (subject !== undefined) {
        const commonNames = subjectAttributeValues(subject, "CN");
        if (commonNames.length > 0) result.subject = { CN: commonNames[0] };
    }
    if (issuerObject !== undefined) result.issuer = issuerObject;
    else if (issuer !== undefined) {
        const commonNames = subjectAttributeValues(issuer, "CN");
        if (commonNames.length > 0) result.issuer = { CN: commonNames[0] };
    }
    if (subjectAltName !== undefined) result.subjectaltname = subjectAltName;
    if (validFrom !== undefined) result.valid_from = validFrom;
    if (validTo !== undefined) result.valid_to = validTo;
    if (fingerprint !== undefined) result.fingerprint = formatFingerprint(fingerprint);
    if (fingerprint256 !== undefined) result.fingerprint256 = formatFingerprint(fingerprint256);
    if (fingerprint512 !== undefined) result.fingerprint512 = formatFingerprint(fingerprint512);
    if (serialNumber !== undefined) result.serialNumber = serialNumber;
    if (ca !== undefined) result.ca = ca;
    if (pem !== undefined) result.pem = pem;
    if (raw !== undefined) result.raw = Buffer.from(hexToBytes(raw));
    else if (pem !== undefined) result.raw = pemToBytes(pem);
    return result;
}

function formatFingerprint(hexValue: string): string {
    const upper = hexValue.toUpperCase();
    const parts: string[] = [];
    for (let i = 0; i < upper.length; i += 2) parts.push(upper.substring(i, i + 2));
    return parts.join(":");
}

function nativeAddress(value: string): AddressInfo {
    const parsed = nativeObject(value);
    return {
        port: parsed.port as number,
        family: parsed.family as string,
        address: parsed.address as string,
    };
}

function nativeCertificate(value: string): NativeCertificate {
    const parsed = nativeObject(value);
    return {
        subject: parsed.subject as string | undefined,
        issuer: parsed.issuer as string | undefined,
        subjectAltName: parsed.subjectAltName as string | undefined,
        validFrom: parsed.validFrom as string | undefined,
        validTo: parsed.validTo as string | undefined,
        fingerprint: parsed.fingerprint as string | undefined,
        fingerprint256: parsed.fingerprint256 as string | undefined,
        fingerprint512: parsed.fingerprint512 as string | undefined,
        serialNumber: parsed.serialNumber as string | undefined,
        pem: parsed.pem as string | undefined,
    };
}

function nativeCipher(value: string): CipherInfo {
    const parsed = nativeObject(value);
    return { name: parsed.name as string, version: parsed.version as string };
}

function nativeEphemeral(value: string): NativeEphemeral {
    const parsed = nativeObject(value);
    return {
        type: parsed.type as string | undefined,
        name: parsed.name as string | undefined,
        size: parsed.size as number | undefined,
        server: parsed.server as boolean | undefined,
    };
}

function nativeCertificateList(value: string): string[] {
    return JSON.parse(value) as string[];
}

export class SecureContext {
    _handle: number;
    _options: SecureContextOptions;

    constructor(options?: SecureContextOptions) {
        const record = optionRecord(options);
        this._options = options === undefined ? {} : options;
        this._handle = __scriptgo.tlsContextCreate(
            pemString(record.cert, "cert"),
            pemString(record.key, "key"),
            optionCA(record),
            optionString(record.minVersion, DEFAULT_MIN_VERSION, "minVersion"),
            optionString(record.maxVersion, DEFAULT_MAX_VERSION, "maxVersion"),
            optionString(record.ciphers, DEFAULT_CIPHERS, "ciphers"),
            optionCAProvided(record),
        );
    }
}

function ensureSecureContext(options?: SecureContextOptions | SecureContext): SecureContext {
    return options instanceof SecureContext ? options : new SecureContext(options);
}

export class X509Certificate {
    subject: string;
    issuer: string;
    subjectAltName: string;
    validFrom: string;
    validTo: string;
    fingerprint: string;
    fingerprint256: string;
    fingerprint512: string;
    serialNumber: string;
    raw: Buffer;
    private _pem: string;

    constructor(bufferOrCert: string | ArrayBufferView) {
        this.subject = "";
        this.issuer = "";
        this.subjectAltName = "";
        this.validFrom = "";
        this.validTo = "";
        this.fingerprint = "";
        this.fingerprint256 = "";
        this.fingerprint512 = "";
        this.serialNumber = "";
        this.raw = Buffer.alloc(0);
        this._pem = "";

        let parsed: NativeCertificate;
        if (typeof bufferOrCert === "string") {
            parsed = nativeCertificate(__scriptgo.tlsX509ParsePem(bufferOrCert));
        } else if (ArrayBuffer.isView(bufferOrCert)) {
            const view = bufferOrCert as ArrayBufferView;
            this.raw = Buffer.from(new Uint8Array(view.buffer, view.byteOffset, view.byteLength));
            parsed = nativeCertificate(__scriptgo.tlsX509ParseBytes(this.raw));
        } else {
            throw new TypeError('The "buffer" argument must be one of type string, Buffer, TypedArray, or DataView');
        }
        this.subject = parsed.subject || "";
        this.issuer = parsed.issuer || "";
        this.subjectAltName = parsed.subjectAltName || "";
        this.validFrom = parsed.validFrom || "";
        this.validTo = parsed.validTo || "";
        this.fingerprint = formatFingerprint(parsed.fingerprint || "");
        this.fingerprint256 = formatFingerprint(parsed.fingerprint256 || "");
        this.fingerprint512 = formatFingerprint(parsed.fingerprint512 || "");
        this.serialNumber = parsed.serialNumber || "";
        this._pem = parsed.pem || "";
        if (typeof bufferOrCert === "string") this.raw = pemToBytes(this._pem);
    }

    checkHost(name: string, options?: X509CheckOptions): string | undefined {
        return matchedHost(name, { subject: this.subject, subjectaltname: this.subjectAltName }, hostCheckOptions(options));
    }

    checkEmail(email: string, options?: X509CheckOptions): string | undefined {
        const subjectOption = subjectCheckOption(options);
        const entries = this.subjectAltName.split(",");
        let hasEmailAlternative = false;
        for (let i = 0; i < entries.length; i++) {
            const entry = entries[i].trim();
            if (entry.substring(0, 6).toLowerCase() !== "email:") continue;
            hasEmailAlternative = true;
            if (matchEmail(entry.substring(6), email)) return email;
        }
        if (subjectOption === "never" || (subjectOption === "default" && hasEmailAlternative)) return undefined;
        const subjectEmails = subjectAttributeValues(this.subject, "emailAddress");
        for (let i = 0; i < subjectEmails.length; i++) {
            if (matchEmail(subjectEmails[i], email)) return email;
        }
        return undefined;
    }

    checkIP(ip: string, options?: X509CheckOptions): string | undefined {
        const entries = this.subjectAltName.split(",");
        let hasIPAlternative = false;
        for (let i = 0; i < entries.length; i++) {
            const entry = entries[i].trim();
            if (entry.startsWith("IP Address:")) {
                hasIPAlternative = true;
                if (entry.substring(11).toLowerCase() === ip.toLowerCase()) return ip;
            }
            if (entry.startsWith("IP:")) {
                hasIPAlternative = true;
                if (entry.substring(3).toLowerCase() === ip.toLowerCase()) return ip;
            }
        }
        const subjectOption = subjectCheckOption(options);
        if (subjectOption === "never" || (subjectOption === "default" && hasIPAlternative)) return undefined;
        const subjectIPs = subjectAttributeValues(this.subject, "CN");
        for (let i = 0; i < subjectIPs.length; i++) {
            if (subjectIPs[i].toLowerCase() === ip.toLowerCase()) return ip;
        }
        return undefined;
    }

    toJSON(): string {
        return this._pem;
    }

    toString(): string {
        return this._pem;
    }
}

type X509HostCheckOptions = {
    subject: X509SubjectOption;
    wildcards: boolean;
    partialWildcards: boolean;
    multiLabelWildcards: boolean;
    singleLabelSubdomains: boolean;
};

function subjectCheckOption(options: { subject?: X509SubjectOption } | undefined): X509SubjectOption {
    if (options === undefined || options.subject === undefined) return "default";
    if (options.subject === "always" || options.subject === "default" || options.subject === "never") return options.subject;
    throw new TypeError("The argument 'options.subject' is invalid");
}

function booleanOption(value: boolean | undefined, fallback: boolean, name: string): boolean {
    if (value === undefined) return fallback;
    if (typeof value !== "boolean") throw new TypeError("The argument 'options." + name + "' must be a boolean");
    return value;
}

function hostCheckOptions(options: X509CheckOptions | undefined): X509HostCheckOptions {
    return {
        subject: subjectCheckOption(options),
        wildcards: booleanOption(options === undefined ? undefined : options.wildcards, true, "wildcards"),
        partialWildcards: booleanOption(options === undefined ? undefined : options.partialWildcards, true, "partialWildcards"),
        multiLabelWildcards: booleanOption(options === undefined ? undefined : options.multiLabelWildcards, false, "multiLabelWildcards"),
        singleLabelSubdomains: booleanOption(options === undefined ? undefined : options.singleLabelSubdomains, false, "singleLabelSubdomains"),
    };
}

function matchEmail(candidate: string, requested: string): boolean {
    const candidateAt = candidate.lastIndexOf("@");
    const requestedAt = requested.lastIndexOf("@");
    if (candidateAt < 0 || requestedAt < 0) return candidate === requested;
    return candidate.substring(0, candidateAt) === requested.substring(0, requestedAt) &&
        candidate.substring(candidateAt).toLowerCase() === requested.substring(requestedAt).toLowerCase();
}

function subjectAttributeValues(subject: string, attribute: string): string[] {
    const values: string[] = [];
    const entries = subject.split("/");
    for (let i = 0; i < entries.length; i++) {
        const entry = entries[i];
        const separator = entry.indexOf("=");
        if (separator < 0 || entry.substring(0, separator).toLowerCase() !== attribute.toLowerCase()) continue;
        values.push(entry.substring(separator + 1));
    }
    return values;
}

function matchHostPattern(pattern: string, host: string, options: X509HostCheckOptions): boolean {
    const normalizedPattern = pattern.toLowerCase();
    const normalizedHost = host.toLowerCase();
    if (normalizedPattern === normalizedHost) return true;

    if (normalizedHost.startsWith(".")) {
        const suffix = normalizedHost.substring(1);
        if (!normalizedPattern.endsWith(suffix) || normalizedPattern.length <= suffix.length || normalizedPattern.charAt(normalizedPattern.length - suffix.length - 1) !== ".") return false;
        const prefix = normalizedPattern.substring(0, normalizedPattern.length - suffix.length - 1);
        if (options.singleLabelSubdomains && prefix.indexOf(".") >= 0) return false;
        return prefix.length > 0;
    }

    if (!options.wildcards) return false;
    const star = normalizedPattern.indexOf("*");
    if (star < 0 || normalizedPattern.indexOf("*", star + 1) >= 0) return false;
    const firstDot = normalizedPattern.indexOf(".");
    if (firstDot < 0 || star > firstDot) return false;
    const prefix = normalizedPattern.substring(0, star);
    const suffix = normalizedPattern.substring(star + 1);
    if (!options.partialWildcards && (prefix.length > 0 || !suffix.startsWith("."))) return false;
    if (prefix.length > 0 && prefix.indexOf(".") >= 0) return false;
    if (suffix.length === 0) return false;
    if (!normalizedHost.endsWith(suffix) || normalizedHost.length <= prefix.length + suffix.length) return false;

    const wildcardValue = normalizedHost.substring(prefix.length, normalizedHost.length - suffix.length);
    if (prefix.length === 0 && suffix.startsWith(".")) {
        if (!options.multiLabelWildcards && wildcardValue.indexOf(".") >= 0) return false;
    } else if (wildcardValue.indexOf(".") >= 0) {
        return false;
    }
    for (let i = 0; i < wildcardValue.length; i++) {
        const ch = wildcardValue.charAt(i);
        const valid = (ch >= "a" && ch <= "z") || (ch >= "0" && ch <= "9") || ch === "-" || (options.multiLabelWildcards && ch === ".");
        if (!valid) return false;
    }
    return wildcardValue.length > 0;
}

function subjectCommonName(subject: string | CertificateName | undefined): string {
    if (typeof subject === "string") {
        const values = subjectAttributeValues(subject, "CN");
        return values.length === 0 ? "" : values[0];
    }
    if (subject !== undefined) {
        const value = subject.CN;
        if (typeof value === "string") return value;
        if (Array.isArray(value) && value.length > 0) return value[0];
    }
    return "";
}

function matchedHost(hostname: string, cert: PeerCertificate, options: X509HostCheckOptions = hostCheckOptions(undefined)): string | undefined {
    const host = hostname.toLowerCase();
    const altName = cert.subjectaltname || "";
    const entries = altName.split(",");
    let hasDNSAlternative = false;
    for (let i = 0; i < entries.length; i++) {
        const entry = entries[i].trim();
        if (entry.startsWith("DNS:")) {
            hasDNSAlternative = true;
            const dns = entry.substring(4).trim();
            if (matchHostPattern(dns, host, options)) return dns;
        }
    }
    if (options.subject === "never" || (options.subject === "default" && hasDNSAlternative)) return undefined;
    const commonName = subjectCommonName(cert.subject);
    return commonName !== "" && matchHostPattern(commonName, host, options) ? commonName : undefined;
}

export function checkServerIdentity(hostname: string, cert: PeerCertificate): Error | undefined {
    return matchedHost(hostname, cert) === undefined ? new Error("Host: " + hostname + ". is not in the cert's altnames") : undefined;
}

export function createSecureContext(options?: SecureContextOptions): SecureContext {
    return new SecureContext(options);
}

export function setDefaultCACertificates(certs: ReadonlyArray<TLSBinary>): void {
    _caCertificates = certificateList(certs, "default CA certificates");
    _caCertificatesSet = true;
}

export function getCACertificates(type?: "default" | "system" | "bundled" | "extra"): string[] {
    const requested = type === undefined ? "default" : type;
    if (requested === "default") {
        if (_caCertificatesSet) return _caCertificates.slice();
        const result = rootCertificates.slice();
        const extra = nativeCertificateList(__scriptgo.tlsExtraCertificates());
        for (let i = 0; i < extra.length; i++) result.push(extra[i]);
        return result;
    }
    if (requested === "bundled") return rootCertificates.slice();
    if (requested === "system") return nativeCertificateList(__scriptgo.tlsSystemCertificates());
    if (requested === "extra") return nativeCertificateList(__scriptgo.tlsExtraCertificates());
    throw new TypeError("The argument 'type' is invalid. Received '" + requested + "'");
}

export function getCiphers(): string[] {
    const ciphers = JSON.parse(__scriptgo.tlsCiphers()) as string[];
    for (let i = 0; i < ciphers.length; i++) ciphers[i] = ciphers[i].toLowerCase();
    return ciphers;
}

export class SecurePair {
    cleartext: TLSSocket;
    encrypted: TLSSocket;
    _buckets: TLSEventBucket[] = [];
    _secureEmitted: boolean = false;

    constructor(context?: SecureContext, isServer?: boolean, requestCert?: boolean, rejectUnauthorized?: boolean) {
        const secureContext = ensureSecureContext(context);
        const handle = __scriptgo.tlsPairCreate(secureContext._handle, isServer === true, requestCert === true, rejectUnauthorized !== false);
        this.cleartext = new TLSSocket(null, { __tlsHandle: handle, __tlsPair: true, __tlsMode: 0 });
        this.encrypted = new TLSSocket(null, { __tlsHandle: handle, __tlsPair: true, __tlsMode: 1 });
        this.cleartext._pairOwner = this;
        this.encrypted._pairOwner = this;
    }

    private _getBucket(name: string): TLSEventBucket {
        for (let i = 0; i < this._buckets.length; i++) if (this._buckets[i].name === name) return this._buckets[i];
        const created = new TLSEventBucket(name, []);
        this._buckets.push(created);
        return created;
    }

    on(event: string, listener: Function): this {
        this._getBucket(event).listeners.push(new TLSEventListenerEntry(listener, false));
        return this;
    }

    once(event: string, listener: Function): this {
        this._getBucket(event).listeners.push(new TLSEventListenerEntry(listener, true));
        return this;
    }

    emit(event: string, arg1: unknown = undefined, arg2: unknown = undefined): boolean {
        const bucket = this._getBucket(event);
        if (bucket.listeners.length === 0) return false;
        const remaining: TLSEventListenerEntry[] = [];
        for (let i = 0; i < bucket.listeners.length; i++) {
            const entry = bucket.listeners[i];
            entry.fn(arg1, arg2);
            if (!entry.once) remaining.push(entry);
        }
        bucket.listeners = remaining;
        return true;
    }

    _markSecure(): void {
        if (this._secureEmitted) return;
        this._secureEmitted = true;
        this.emit("secure");
    }
}

export class Server {
    listening: boolean = false;
    maxConnections: number = 1000;
    dropMaxConnection: boolean = false;
    maxHeadersCount: number = 2000;
    timeout: number = 0;
    keepAliveTimeout: number = 5000;
    _connectionsCount: number = 0;
    _secureContext: SecureContext;
    _ticketKeys: Buffer;
    _handle: number = 0;
    _buckets: TLSEventBucket[] = [];
    _sniContexts: Map<string, SecureContext> = new Map();
    _requestCert: boolean = false;
    _rejectUnauthorized: boolean = true;
    _host: string = "0.0.0.0";
    _addressPort: number = 0;

    constructor(options?: TLSServerOptions | TLSSecureConnectionListener, listener?: TLSSecureConnectionListener) {
        const optionValue = typeof options === "function" ? undefined : options;
        const record = optionRecord(optionValue);
        this._secureContext = new SecureContext(optionValue);
        this._requestCert = record.requestCert === true;
        this._rejectUnauthorized = record.rejectUnauthorized !== false;
        this._ticketKeys = Buffer.from(__scriptgo.randomBytes(48));
        if (typeof options === "function") this.on("secureConnection", options as Function);
        else if (typeof listener === "function") this.on("secureConnection", listener as Function);
    }

    private _getBucket(name: string): TLSEventBucket {
        for (let i = 0; i < this._buckets.length; i++) if (this._buckets[i].name === name) return this._buckets[i];
        const created = new TLSEventBucket(name, []);
        this._buckets.push(created);
        return created;
    }

    on(event: string, listener: Function): this {
        this._getBucket(event).listeners.push(new TLSEventListenerEntry(listener, false));
        return this;
    }

    once(event: string, listener: Function): this {
        this._getBucket(event).listeners.push(new TLSEventListenerEntry(listener, true));
        return this;
    }

    emit(event: string, arg1: unknown = undefined, arg2: unknown = undefined): boolean {
        const bucket = this._getBucket(event);
        if (bucket.listeners.length === 0) return false;
        const remaining: TLSEventListenerEntry[] = [];
        for (let i = 0; i < bucket.listeners.length; i++) {
            const entry = bucket.listeners[i];
            entry.fn(arg1, arg2);
            if (!entry.once) remaining.push(entry);
        }
        bucket.listeners = remaining;
        return true;
    }

    addContext(hostname: string, context: SecureContextOptions | SecureContext): void {
        const secureContext = ensureSecureContext(context);
        this._sniContexts.set(hostname, secureContext);
        if (this._handle !== 0) __scriptgo.tlsServerAddContext(this._handle, hostname, secureContext._handle);
    }

    address(): AddressInfo | null {
        if (!this.listening || this._handle === 0) return null;
        return nativeAddress(__scriptgo.tlsServerInfo(this._handle, "address"));
    }

    accept(): TLSSocket {
        if (this._handle === 0) throw new Error("TLS server is not listening");
        const socket = new TLSSocket(null, { __tlsHandle: __scriptgo.tlsServerAccept(this._handle), __tlsPair: false, __tlsMode: 0, __tlsConnected: true });
        this._connectionsCount++;
        this.emit("secureConnection", socket);
        return socket;
    }

    close(callback?: TLSCallback): this {
        if (callback !== undefined) this.once("close", callback);
        if (this._handle !== 0) {
            __scriptgo.tlsServerClose(this._handle);
            this._handle = 0;
        }
        this.listening = false;
        this.emit("close");
        return this;
    }

    getTicketKeys(): Buffer {
        return Buffer.from(this._ticketKeys);
    }

    listen(portOrOptions: number | TLSListenOptions = 0, hostOrCb?: string | TLSCallback, callback?: TLSCallback): this {
        let listenPort = 0;
        let listenHost = "0.0.0.0";
        let backlog = 511;
        let cb: TLSCallback | undefined = callback;
        if (typeof portOrOptions === "number") {
            listenPort = portOrOptions;
            if (typeof hostOrCb === "string") listenHost = hostOrCb;
            else if (typeof hostOrCb === "function") cb = hostOrCb;
        } else {
            const options = portOrOptions;
            if (options.port !== undefined) listenPort = options.port;
            if (options.host !== undefined) listenHost = options.host;
            if (options.backlog !== undefined) backlog = options.backlog;
            if (typeof hostOrCb === "function") cb = hostOrCb;
        }
        if (cb !== undefined) this.once("listening", cb);
        this._host = listenHost;
        this._handle = __scriptgo.tlsServerListen(this._secureContext._handle, this._requestCert, this._rejectUnauthorized, listenHost, listenPort, backlog);
        for (const [hostname, context] of this._sniContexts) __scriptgo.tlsServerAddContext(this._handle, hostname, context._handle);
        __scriptgo.tlsServerSetTicketKeys(this._handle, bytesToHex(this._ticketKeys));
        this.listening = true;
        const info = nativeAddress(__scriptgo.tlsServerInfo(this._handle, "address"));
        this._addressPort = info.port;
        this.emit("listening");
        return this;
    }

    setSecureContext(options: SecureContextOptions): void {
        this._secureContext = new SecureContext(options);
        if (this._handle !== 0) __scriptgo.tlsServerSetContext(this._handle, this._secureContext._handle, this._requestCert, this._rejectUnauthorized);
    }

    setTicketKeys(keys: Uint8Array): void {
        const value = byteView(keys, "ticket keys", false);
        if (value.length !== 48) throw new RangeError("TLS ticket keys must contain exactly 48 bytes");
        this._ticketKeys = Buffer.from(value);
        if (this._handle !== 0) __scriptgo.tlsServerSetTicketKeys(this._handle, bytesToHex(this._ticketKeys));
    }
}

export function createServer(options?: TLSServerOptions | TLSSecureConnectionListener, secureConnectionListener?: TLSSecureConnectionListener): Server {
    return new Server(options, secureConnectionListener);
}

export class TLSSocket {
    authorizationError: string | null = null;
    authorized: boolean = false;
    encrypted: boolean = true;
    localAddress: string | undefined = undefined;
    localPort: number | undefined = undefined;
    localFamily: string | undefined = undefined;
    remoteAddress: string | undefined = undefined;
    remoteFamily: string | undefined = undefined;
    remotePort: number | undefined = undefined;
    _underlyingSocket: Socket | Duplex | null = null;
    _tlsOptions: TLSSocketOptions | InternalTLSSocketOptions | null = null;
    _handle: number = 0;
    _isPair: boolean = false;
    _pairMode: number = 0;
    _connected: boolean = false;
    _closed: boolean = false;
    _buckets: TLSEventBucket[] = [];
    _renegotiationDisabled: boolean = false;
    _traceEnabled: boolean = false;
    _maxSendFragment: number = 16384;
    _servername: string = "";
    _session: Uint8Array | undefined = undefined;
    _pairOwner: SecurePair | undefined = undefined;

    constructor(socket?: Socket | Duplex | null, options?: TLSSocketOptions | InternalTLSSocketOptions) {
        this._underlyingSocket = socket || null;
        this._tlsOptions = options === undefined ? null : options;
        const internal = optionRecord(options) as InternalTLSSocketOptions;
        this._servername = internal.servername || "";
        if (internal.session !== undefined) {
            this._session = byteView(internal.session, "session", false);
        }
        if (internal.__tlsHandle !== undefined) {
            this._handle = internal.__tlsHandle;
            this._isPair = internal.__tlsPair === true;
            this._pairMode = internal.__tlsMode || 0;
            this._connected = internal.__tlsConnected === true;
            this._refreshState();
            return;
        }
        if (socket === null || socket === undefined) {
            const record = optionRecord(options);
            const context = ensureSecureContext(record.secureContext === undefined ? record : record.secureContext);
            this._handle = __scriptgo.tlsSocketCreate(context._handle, record.isServer === true);
            if (this._servername !== "") __scriptgo.tlsSocketSetServername(this._handle, this._servername);
            if (this._session !== undefined) __scriptgo.tlsSocketSetSession(this._handle, this._session);
            return;
        }
        if (socket !== null && socket !== undefined && typeof socket === "object") {
            const source = socket as Socket;
            const fd = source._fd;
            if (fd >= 0) {
                const record = optionRecord(options);
                const context = ensureSecureContext(record.secureContext === undefined ? record : record.secureContext);
                this._handle = __scriptgo.tlsSocketAdopt(context._handle, fd, record.servername || "", record.isServer === true, record.requestCert === true, record.rejectUnauthorized !== false, this._session === undefined ? new Uint8Array(0) : this._session);
                this._connected = true;
                this._refreshState();
            }
        }
    }

    private _getBucket(name: string): TLSEventBucket {
        for (let i = 0; i < this._buckets.length; i++) if (this._buckets[i].name === name) return this._buckets[i];
        const created = new TLSEventBucket(name, []);
        this._buckets.push(created);
        return created;
    }

    private _refreshState(): void {
        if (this._handle === 0) return;
        const localAddress = __scriptgo.tlsSocketInfo(this._handle, "localAddress");
        const localFamily = __scriptgo.tlsSocketInfo(this._handle, "localFamily");
        const remoteAddress = __scriptgo.tlsSocketInfo(this._handle, "remoteAddress");
        const remoteFamily = __scriptgo.tlsSocketInfo(this._handle, "remoteFamily");
        const localPort = __scriptgo.tlsSocketNumber(this._handle, "localPort");
        const remotePort = __scriptgo.tlsSocketNumber(this._handle, "remotePort");
        this.localAddress = localAddress === "" ? undefined : localAddress;
        this.localFamily = localFamily === "" ? undefined : localFamily;
        this.remoteAddress = remoteAddress === "" ? undefined : remoteAddress;
        this.remoteFamily = remoteFamily === "" ? undefined : remoteFamily;
        this.localPort = localPort === 0 ? undefined : localPort;
        this.remotePort = remotePort === 0 ? undefined : remotePort;
        const error = __scriptgo.tlsSocketInfo(this._handle, "authorizationError");
        this.authorizationError = error.length === 0 ? null : error;
        this.authorized = __scriptgo.tlsSocketBool(this._handle, "authorized");
    }

    private _applyOptions(): void {
        if (this._handle === 0 || this._isPair) return;
        if (this._renegotiationDisabled) __scriptgo.tlsSocketSetOption(this._handle, "disableRenegotiation", 1);
        if (this._traceEnabled) __scriptgo.tlsSocketSetOption(this._handle, "trace", 1);
        if (this._maxSendFragment !== 16384) __scriptgo.tlsSocketSetOption(this._handle, "maxSendFragment", this._maxSendFragment);
    }

    private _notifyPairSecure(): void {
        if (!this._isPair || this._pairOwner === undefined || this._handle === 0) return;
        this._refreshState();
        if (__scriptgo.tlsSocketBool(this._handle, "handshakeFinished")) {
            this._connected = true;
            this._pairOwner._markSecure();
        }
    }

    on(event: string, listener: Function): this {
        this._getBucket(event).listeners.push(new TLSEventListenerEntry(listener, false));
        return this;
    }

    once(event: string, listener: Function): this {
        this._getBucket(event).listeners.push(new TLSEventListenerEntry(listener, true));
        return this;
    }

    emit(event: string, arg1: unknown = undefined, arg2: unknown = undefined): boolean {
        const bucket = this._getBucket(event);
        if (bucket.listeners.length === 0) return false;
        const remaining: TLSEventListenerEntry[] = [];
        for (let i = 0; i < bucket.listeners.length; i++) {
            const entry = bucket.listeners[i];
            entry.fn(arg1, arg2);
            if (!entry.once) remaining.push(entry);
        }
        bucket.listeners = remaining;
        return true;
    }

    connect(optionsOrPort: number | TLSConnectionOptions, hostOrOptionsOrCallback?: string | TLSConnectionOptions | TLSCallback, optionsOrCallback?: TLSConnectionOptions | TLSCallback, callback?: TLSCallback): this {
        let port = 0;
        let host = "localhost";
        let options: TLSOptionRecord = {};
        let secureConnectListener: TLSCallback | undefined;
        if (typeof optionsOrPort === "number") {
            port = optionsOrPort;
            if (typeof hostOrOptionsOrCallback === "string") {
                host = hostOrOptionsOrCallback;
                if (typeof optionsOrCallback === "object") options = optionRecord(optionsOrCallback);
                else if (typeof optionsOrCallback === "function") secureConnectListener = optionsOrCallback;
            } else if (typeof hostOrOptionsOrCallback === "object") {
                options = optionRecord(hostOrOptionsOrCallback);
            } else if (typeof hostOrOptionsOrCallback === "function") {
                secureConnectListener = hostOrOptionsOrCallback;
            }
            if (typeof optionsOrCallback === "function") secureConnectListener = optionsOrCallback;
            if (callback !== undefined) secureConnectListener = callback;
        } else {
            options = optionRecord(optionsOrPort);
            if (options.host !== undefined) host = options.host;
            if (options.port !== undefined) port = options.port;
            if (typeof hostOrOptionsOrCallback === "function") secureConnectListener = hostOrOptionsOrCallback;
            if (typeof optionsOrCallback === "function") secureConnectListener = optionsOrCallback;
            if (callback !== undefined) secureConnectListener = callback;
        }
        if (secureConnectListener !== undefined) this.once("secureConnect", secureConnectListener);
        if (this._handle !== 0) {
            __scriptgo.tlsSocketClose(this._handle);
            this._handle = 0;
        }
        this._connected = false;
        const context = ensureSecureContext(options.secureContext === undefined ? options : options.secureContext);
        if (options.servername !== undefined) this._servername = options.servername;
        if (options.session !== undefined) {
            this._session = byteView(options.session, "session", false);
        }
        this._handle = __scriptgo.tlsSocketConnect(context._handle, host, port, this._servername, options.rejectUnauthorized !== false, this._session === undefined ? new Uint8Array(0) : this._session);
        this._isPair = false;
        this._connected = true;
        this._closed = false;
        this._applyOptions();
        this._refreshState();
        this.emit("secureConnect");
        return this;
    }

    write(data: string | Uint8Array, callback?: TLSCallback): boolean {
        if (this._handle === 0 || this._closed) throw new Error("TLS socket is not open");
        if (typeof data === "string") {
            const length = this._isPair && this._pairMode === 1 ? data.length : Buffer.byteLength(data, "utf8");
            if (this._isPair) __scriptgo.tlsPairWrite(this._handle, this._pairMode, data, length);
            else __scriptgo.tlsSocketWrite(this._handle, data, length);
        } else if (this._isPair) {
            __scriptgo.tlsPairWriteBytes(this._handle, this._pairMode, data);
        } else {
            __scriptgo.tlsSocketWriteBytes(this._handle, data);
        }
        this._notifyPairSecure();
        if (callback !== undefined) callback();
        return true;
    }

    read(size: number = 65536): string {
        if (this._handle === 0 || this._closed) return "";
        const data = this._isPair ? __scriptgo.tlsPairRead(this._handle, this._pairMode, size) : __scriptgo.tlsSocketRead(this._handle, size);
        this._notifyPairSecure();
        if (data.length === 0) this.emit("end");
        else this.emit("data", data);
        return data;
    }

    end(data?: string | Uint8Array, callback?: TLSCallback): this {
        if (data !== undefined) this.write(data);
        if (callback !== undefined) this.once("finish", callback);
        this.emit("finish");
        this.destroy();
        return this;
    }

    address(): SocketAddressInfo {
        if (this._handle === 0 || this.localAddress === undefined || this.localFamily === undefined || this.localPort === undefined) return {};
        return { port: this.localPort, family: this.localFamily, address: this.localAddress };
    }

    disableRenegotiation(): void {
        this._renegotiationDisabled = true;
        if (this._handle !== 0 && !this._isPair) __scriptgo.tlsSocketSetOption(this._handle, "disableRenegotiation", 1);
    }

    enableTrace(): void {
        this._traceEnabled = true;
        if (this._handle !== 0 && !this._isPair) __scriptgo.tlsSocketSetOption(this._handle, "trace", 1);
    }

    exportKeyingMaterial(length: number, label: string, context?: Uint8Array): Uint8Array {
        if (this._handle === 0) throw new Error("TLS socket is not connected");
        const contextBytes = context === undefined ? null : byteView(context, "export keying material context");
        return hexToBytes(__scriptgo.tlsExportKeyingMaterial(this._handle, length, label, contextBytes));
    }

    getCertificate(): PeerCertificate | null {
        if (this._closed || this._handle === 0) return null;
        return nativeTLSCertificate(__scriptgo.tlsSocketInfo(this._handle, "certificate"));
    }

    getCipher(): CipherInfo | null | undefined {
        if (this._closed) return null;
        if (this._handle === 0 || !this._connected) return undefined;
        return nativeCipher(__scriptgo.tlsSocketInfo(this._handle, "cipher"));
    }

    getEphemeralKeyInfo(): EphemeralKeyInfo | null {
        if (this._closed || this._handle === 0) return null;
        const value = nativeEphemeral(__scriptgo.tlsSocketInfo(this._handle, "ephemeral"));
        if (value.server === true) return null;
        if (value.type === undefined && value.name === undefined && value.size === undefined) return new EphemeralKeyInfo();
        const result = new EphemeralKeyInfo();
        result.type = value.type || undefined;
        result.name = value.name || undefined;
        result.size = value.size || undefined;
        return result;
    }

    getFinished(): Uint8Array | null | undefined {
        if (this._closed) return null;
        if (this._handle === 0 || !this._connected) return undefined;
        return hexToBytes(__scriptgo.tlsSocketInfo(this._handle, "finished"));
    }

    getPeerCertificate(detailed?: boolean): PeerCertificate | null {
        if (this._closed || this._handle === 0) return null;
        return nativeTLSCertificate(__scriptgo.tlsSocketInfo(this._handle, detailed === true ? "peerCertificateDetailed" : "peerCertificate"));
    }

    getPeerFinished(): Uint8Array | null | undefined {
        if (this._closed) return null;
        if (this._handle === 0 || !this._connected) return undefined;
        return hexToBytes(__scriptgo.tlsSocketInfo(this._handle, "peerFinished"));
    }

    getPeerX509Certificate(): X509Certificate | undefined {
        if (this._handle === 0 || !this._connected) return undefined;
        const value = nativeCertificate(__scriptgo.tlsSocketInfo(this._handle, "peerCertificateDetailed"));
        const pem = value.pem;
        return pem === undefined || pem === "" ? undefined : new X509Certificate(pem);
    }

    getProtocol(): string | null {
        if (this._closed || this._handle === 0) return null;
        return __scriptgo.tlsSocketInfo(this._handle, "protocol");
    }

    getSession(): Uint8Array | null | undefined {
        if (this._closed) return null;
        if (this._handle === 0 || !this._connected) return undefined;
        const value = __scriptgo.tlsSocketInfo(this._handle, "session");
        return value.length === 0 ? undefined : hexToBytes(value);
    }

    getSharedSigalgs(): string[] | null {
        if (this._closed) return null;
        if (this._handle === 0 || !this._connected) return [];
        return JSON.parse(__scriptgo.tlsSocketInfo(this._handle, "sharedSigalgs")) as string[];
    }

    getTLSTicket(): Uint8Array | null | undefined {
        if (this._closed) return null;
        if (this._handle === 0 || !this._connected) return undefined;
        const value = __scriptgo.tlsSocketInfo(this._handle, "ticket");
        return value.length === 0 ? undefined : hexToBytes(value);
    }

    getX509Certificate(): X509Certificate | undefined {
        if (this._closed || this._handle === 0) return undefined;
        const value = nativeCertificate(__scriptgo.tlsSocketInfo(this._handle, "certificate"));
        const pem = value.pem;
        return pem === undefined || pem === "" ? undefined : new X509Certificate(pem);
    }

    isSessionReused(): boolean | null {
        if (this._closed) return null;
        return __scriptgo.tlsSocketBool(this._handle, "sessionReused");
    }

    renegotiate(options: { requestCert?: boolean; rejectUnauthorized?: boolean }, callback: TLSRenegotiationCallback): boolean {
        const record = optionRecord(options);
        if (this._handle !== 0 && !this._isPair) {
            if (record.requestCert !== undefined) __scriptgo.tlsSocketSetOption(this._handle, "requestCert", record.requestCert === true ? 1 : 0);
            if (record.rejectUnauthorized !== undefined) __scriptgo.tlsSocketSetOption(this._handle, "rejectUnauthorized", record.rejectUnauthorized === true ? 1 : 0);
        }
        const result = this._handle !== 0 && !this._isPair && __scriptgo.tlsSocketRenegotiate(this._handle);
        callback(result ? null : new Error("TLS renegotiation was not initiated"));
        return result;
    }

    setKeyCert(options: SecureContextOptions | SecureContext): void {
        if (this._handle === 0 || this._isPair) throw new Error("TLS socket is not open");
        const record = optionRecord(options instanceof SecureContext ? options._options : options);
        __scriptgo.tlsSocketSetKeyCert(this._handle, pemString(record.cert, "cert"), pemString(record.key, "key"));
    }

    setMaxSendFragment(size: number): boolean {
        if (size !== size || size !== Math.floor(size)) throw new RangeError("TLS max send fragment must be an integer");
        if (size < 512 || size > 16384) return false;
        this._maxSendFragment = size;
        if (this._handle !== 0 && !this._isPair) __scriptgo.tlsSocketSetOption(this._handle, "maxSendFragment", size);
        return true;
    }

    setServername(servername: string): void {
        if (this._handle === 0 || this._closed) throw new Error("TLS socket is not open");
        if (this._connected) throw new Error("TLS socket handshake has already started");
        this._servername = servername;
        __scriptgo.tlsSocketSetServername(this._handle, servername);
    }

    setSession(session: Uint8Array): void {
        if (this._handle === 0 || this._closed) throw new Error("TLS socket is not open");
        if (this._connected) throw new Error("TLS socket handshake has already started");
        if (!(session instanceof Uint8Array) || session.length === 0) throw new TypeError("TLS session must be a non-empty Uint8Array");
        this._session = session.slice();
        __scriptgo.tlsSocketSetSession(this._handle, this._session);
    }

    destroy(error?: Error): this {
        if (this._closed) return this;
        if (this._handle !== 0) {
            __scriptgo.tlsSocketClose(this._handle);
            this._handle = 0;
        }
        this._connected = false;
        this._closed = true;
        if (error !== undefined) this.emit("error", error);
        this.emit("close", error !== undefined);
        return this;
    }
}

export function createSecurePair(context?: SecureContext, isServer?: boolean, requestCert?: boolean, rejectUnauthorized?: boolean): SecurePair {
    return new SecurePair(context, isServer, requestCert, rejectUnauthorized);
}

export function connect(optionsOrPort: number | TLSConnectionOptions, hostOrOptionsOrCallback?: string | TLSConnectionOptions | TLSCallback, optionsOrCallback?: TLSConnectionOptions | TLSCallback, callback?: TLSCallback): TLSSocket {
    const socket = new TLSSocket();
    if (typeof optionsOrPort === "number") {
        if (typeof hostOrOptionsOrCallback === "string") {
            if (typeof optionsOrCallback === "object") {
                socket.connect(optionsOrPort, hostOrOptionsOrCallback, optionsOrCallback, callback);
            } else if (typeof optionsOrCallback === "function") {
                socket.connect(optionsOrPort, hostOrOptionsOrCallback, optionsOrCallback);
            } else if (callback !== undefined) {
                socket.connect(optionsOrPort, hostOrOptionsOrCallback, undefined, callback);
            } else {
                socket.connect(optionsOrPort, hostOrOptionsOrCallback);
            }
        } else if (typeof hostOrOptionsOrCallback === "object") {
            if (typeof optionsOrCallback === "function") socket.connect(optionsOrPort, hostOrOptionsOrCallback, optionsOrCallback);
            else socket.connect(optionsOrPort, hostOrOptionsOrCallback);
        } else if (typeof hostOrOptionsOrCallback === "function") {
            socket.connect(optionsOrPort, hostOrOptionsOrCallback);
        } else {
            socket.connect(optionsOrPort);
        }
    } else if (typeof hostOrOptionsOrCallback === "function") {
        socket.connect(optionsOrPort, hostOrOptionsOrCallback);
    } else {
        socket.connect(optionsOrPort);
    }
    return socket;
}

export default {
    CLIENT_RENEG_LIMIT,
    CLIENT_RENEG_WINDOW,
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
    SecureContext,
    SecurePair,
    createSecurePair,
    Server,
    createServer,
    TLSSocket,
    connect,
    X509Certificate,
};
