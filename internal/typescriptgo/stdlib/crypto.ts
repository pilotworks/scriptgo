import { EventEmitter } from "node:events";
import { SubtleCrypto, Crypto, webcrypto } from "webcrypto";

// These calls are lowered to the linked native crypto runtime. Keeping the
// adapter here makes the public Node-shaped classes use the same ABI as the
// promoted module functions.
declare namespace __scriptgo {
    function hashDigest(algorithm: string, data: string, encoding?: string): string;
    function hashDigestBuffer(algorithm: string, data: Buffer, encoding?: string): string;
    function hmacDigest(algorithm: string, key: string, data: string, encoding?: string): string;
    function hmacDigestBuffer(algorithm: string, key: Buffer, data: Buffer, encoding?: string): string;
    function randomUUID(): string;
    function randomBytes(size: number): Buffer;
    function randomInt(min: number, max: number): number;
    function randomFill(buffer: Buffer, offset?: number, size?: number): Buffer;
    function timingSafeEqual(a: Buffer, b: Buffer): boolean;
    function pbkdf2Sync(password: string, salt: string, iterations: number, keylen: number, digest?: string): Buffer;
    function hkdfSync(digest: string, ikm: string, salt: string, info: string, keylen: number): ArrayBuffer;
    function scryptSync(password: string, salt: string, keylen: number): Buffer;
}

type CryptoBinary = string | Buffer | Uint8Array | ArrayBuffer;

function toCryptoBuffer(data: CryptoBinary, encoding?: string): Buffer {
    return typeof data === "string" ? Buffer.from(data, encoding) : Buffer.from(data);
}

export interface HashOptions {
    outputLength?: number;
}

export class Hash extends EventEmitter {
    private _algorithm: string;
    private _data: Buffer;

    constructor(algorithm: string = "sha256") {
        super();
        this._algorithm = algorithm;
        this._data = Buffer.alloc(0);
    }

    update(data: CryptoBinary, inputEncoding?: string): this {
        this._data = Buffer.concat([this._data, toCryptoBuffer(data, inputEncoding)]);
        return this;
    }

    digest(encoding: string): string;
    digest(): Buffer;
    digest(encoding?: string): Buffer | string {
        const hex = __scriptgo.hashDigestBuffer(this._algorithm, this._data, "hex");
        if (encoding !== undefined) {
            return __scriptgo.hashDigestBuffer(this._algorithm, this._data, encoding);
        }
        return Buffer.from(hex, "hex");
    }

    copy(options?: HashOptions): Hash {
        const copy = new Hash(this._algorithm);
        copy._data = this._data;
        return copy;
    }
}

export class Hmac extends EventEmitter {
    private _algorithm: string;
    private _key: Buffer;
    private _data: Buffer;

    constructor(algorithm: string = "sha256", key: CryptoBinary = "") {
        super();
        this._algorithm = algorithm;
        this._key = toCryptoBuffer(key);
        this._data = Buffer.alloc(0);
    }

    update(data: CryptoBinary, inputEncoding?: string): this {
        this._data = Buffer.concat([this._data, toCryptoBuffer(data, inputEncoding)]);
        return this;
    }

    digest(encoding: string): string;
    digest(): Buffer;
    digest(encoding?: string): Buffer | string {
        const hex = __scriptgo.hmacDigestBuffer(this._algorithm, this._key, this._data, "hex");
        if (encoding !== undefined) {
            return __scriptgo.hmacDigestBuffer(this._algorithm, this._key, this._data, encoding);
        }
        return Buffer.from(hex, "hex");
    }

    copy(options?: HashOptions): Hmac {
        const copy = new Hmac(this._algorithm, this._key);
        copy._data = this._data;
        return copy;
    }
}

function formatFingerprintCrypto(hexStr: string): string {
    const upper = hexStr.toUpperCase();
    const parts: string[] = [];
    for (let i = 0; i < upper.length; i += 2) {
        parts.push(upper.substring(i, i + 2));
    }
    return parts.join(":");
}

export class X509Certificate {
    ca: boolean = false;
    fingerprint: string = "";
    fingerprint256: string = "";
    fingerprint512: string = "";
    infoAccess: string = "";
    issuer: string = "";
    issuerCertificate: X509Certificate | undefined = undefined;
    keyUsage: string[] = [];
    publicKey: Record<string, unknown> = { type: "public" };
    raw: Buffer = Buffer.alloc(0);
    serialNumber: string = "";
    subject: string = "";
    subjectAltName: string = "";
    validFrom: string = "";
    validFromDate: string = "";
    validTo: string = "";
    validToDate: string = "";
    private _pem: string = "";

    constructor(bufferOrCert: unknown) {
        if (bufferOrCert === undefined || bufferOrCert === null) {
            throw new TypeError('The "buffer" argument must be one of type string, Buffer, TypedArray, or DataView');
        }

        let certStr = "";
        if (typeof bufferOrCert === "string") {
            certStr = bufferOrCert as string;
            this._pem = certStr;
            let b64 = certStr;
            const beginIdx = certStr.indexOf("-----BEGIN CERTIFICATE-----");
            const endIdx = certStr.indexOf("-----END CERTIFICATE-----");
            if (beginIdx >= 0 && endIdx >= 0) {
                b64 = certStr.substring(beginIdx + 27, endIdx).replace(/\s+/g, "");
            } else if (!certStr.startsWith("Subject:") && !certStr.startsWith("subject=")) {
                throw new Error("error:0480006C:PEM routines::no start line");
            }
            this.raw = Buffer.from(b64, "base64");
        } else if (bufferOrCert instanceof Uint8Array || Buffer.isBuffer(bufferOrCert)) {
            this.raw = Buffer.from(bufferOrCert as Uint8Array);
        } else {
            throw new TypeError('The "buffer" argument must be one of type string, Buffer, TypedArray, or DataView');
        }

        try {
            const sha1Hex = __scriptgo.hashDigestBuffer("sha1", this.raw, "hex");
            this.fingerprint = formatFingerprintCrypto(sha1Hex);
            const sha256Hex = __scriptgo.hashDigestBuffer("sha256", this.raw, "hex");
            this.fingerprint256 = formatFingerprintCrypto(sha256Hex);
            const sha512Hex = __scriptgo.hashDigestBuffer("sha512", this.raw, "hex");
            this.fingerprint512 = formatFingerprintCrypto(sha512Hex);
        } catch {
            this.fingerprint = "";
            this.fingerprint256 = "";
            this.fingerprint512 = "";
        }

        if (this.raw && this.raw.length > 5) {
            for (let i = 0; i < this.raw.length - 5; i++) {
                if (this.raw[i] === 0x55 && this.raw[i + 1] === 0x04 && this.raw[i + 2] === 0x03) {
                    const strLen = this.raw[i + 4];
                    if (i + 5 + strLen <= this.raw.length) {
                        this.subject = "CN=" + this.raw.subarray(i + 5, i + 5 + strLen).toString("utf8");
                        break;
                    }
                }
            }
        }

        if (certStr.length > 0) {
            const lines = certStr.split("\n");
            for (let i = 0; i < lines.length; i++) {
                const line = lines[i].trim();
                if (line.startsWith("Subject:") || line.startsWith("subject=")) {
                    this.subject = line.substring(line.indexOf(":") >= 0 ? line.indexOf(":") + 1 : line.indexOf("=") + 1).trim();
                } else if (line.startsWith("Issuer:") || line.startsWith("issuer=")) {
                    this.issuer = line.substring(line.indexOf(":") >= 0 ? line.indexOf(":") + 1 : line.indexOf("=") + 1).trim();
                } else if (line.startsWith("DNS:") || line.startsWith("IP Address:") || line.startsWith("SAN:")) {
                    this.subjectAltName = line;
                } else if (line.startsWith("Not Before:") || line.startsWith("validFrom=")) {
                    this.validFrom = line.substring(line.indexOf(":") >= 0 ? line.indexOf(":") + 1 : line.indexOf("=") + 1).trim();
                    this.validFromDate = this.validFrom;
                } else if (line.startsWith("Not After :") || line.startsWith("Not After:") || line.startsWith("validTo=")) {
                    this.validTo = line.substring(line.indexOf(":") >= 0 ? line.indexOf(":") + 1 : line.indexOf("=") + 1).trim();
                    this.validToDate = this.validTo;
                } else if (line.startsWith("Serial Number:") || line.startsWith("serial=")) {
                    this.serialNumber = line.substring(line.indexOf(":") >= 0 ? line.indexOf(":") + 1 : line.indexOf("=") + 1).trim();
                }
            }
        }
        if (this.issuer === "") {
            this.issuer = this.subject;
        }
        if (this.serialNumber === "") {
            this.serialNumber = "01";
        }
    }

    checkEmail(email: string, options?: unknown): boolean {
        return this.subjectAltName.indexOf("email:" + email) >= 0 || this.subject.indexOf("emailAddress=" + email) >= 0;
    }

    checkHost(host: string, options?: unknown): string | undefined {
        if (this.subjectAltName.indexOf("DNS:" + host) >= 0 || this.subject.indexOf("CN=" + host) >= 0) {
            return undefined;
        }
        return host;
    }

    checkIP(ip: string, options?: unknown): string | undefined {
        if (this.subjectAltName.indexOf("IP Address:" + ip) >= 0 || this.subjectAltName.indexOf("IP:" + ip) >= 0) {
            return undefined;
        }
        return ip;
    }

    checkIssued(otherCert: X509Certificate): boolean {
        return this.subject === otherCert.issuer;
    }

    toJSON(): string {
        return JSON.stringify({
            subject: this.subject,
            issuer: this.issuer,
            subjectAltName: this.subjectAltName,
            validFrom: this.validFrom,
            validTo: this.validTo,
            fingerprint: this.fingerprint,
            fingerprint256: this.fingerprint256,
            fingerprint512: this.fingerprint512,
            serialNumber: this.serialNumber,
        });
    }

    toLegacyObject(): unknown {
        return {
            subject: this.subject,
            issuer: this.issuer,
            valid_from: this.validFrom,
            valid_to: this.validTo,
            fingerprint: this.fingerprint,
            fingerprint256: this.fingerprint256,
            fingerprint512: this.fingerprint512,
            serialNumber: this.serialNumber,
        };
    }

    toString(): string {
        return this._pem.length > 0 ? this._pem : "-----BEGIN CERTIFICATE-----\n" + this.fingerprint256 + "\n-----END CERTIFICATE-----";
    }
}

export const constants: Record<string, number> = {
    RSA_PKCS1_PADDING: 1,
    RSA_SSLV23_PADDING: 2,
    RSA_NO_PADDING: 3,
    RSA_PKCS1_OAEP_PADDING: 4,
    RSA_X931_PADDING: 5,
    RSA_PKCS1_PSS_PADDING: 6,
    POINT_CONVERSION_COMPRESSED: 2,
    POINT_CONVERSION_UNCOMPRESSED: 4,
    POINT_CONVERSION_HYBRID: 6,
};

export { webcrypto };
export const subtle: SubtleCrypto = webcrypto.subtle;

export function checkPrime(candidate: unknown, callback: (err: Error | null, result: boolean) => void): void;
export function checkPrime(candidate: unknown, options: unknown, callback: (err: Error | null, result: boolean) => void): void;
export function checkPrime(candidate: unknown, options?: unknown, callback?: unknown): void {
    const result = checkPrimeSync(candidate, options);
    if (typeof options === "function") {
        (options as (err: Error | null, result: boolean) => void)(null, result);
    } else if (typeof callback === "function") {
        (callback as (err: Error | null, result: boolean) => void)(null, result);
    }
}

export function checkPrimeSync(candidate: unknown, options?: unknown): boolean {
	if (typeof candidate === "bigint") {
		if (candidate < 2n) return false;
		if (candidate === 2n || candidate === 3n) return true;
		if (candidate % 2n === 0n) return false;
		for (let divisor = 3n; divisor * divisor <= candidate; divisor += 2n) {
			if (candidate % divisor === 0n) return false;
		}
		return true;
	}
	if (typeof candidate !== "number" || !Number.isSafeInteger(candidate) || candidate < 2) {
        return false;
    }
    if (candidate === 2 || candidate === 3) {
        return true;
    }
    if (candidate % 2 === 0) {
        return false;
    }
    for (let divisor = 3; divisor * divisor <= candidate; divisor += 2) {
        if (candidate % divisor === 0) {
            return false;
        }
    }
    return true;
}

export function createHash(algorithm: string, options?: unknown): Hash {
    return new Hash(algorithm);
}

export function createHmac(algorithm: string, key: CryptoBinary, options?: unknown): Hmac {
    return new Hmac(algorithm, key);
}

export function generatePrime(size: number, callback: (err: Error | null, prime: unknown) => void): void;
export function generatePrime(size: number, options: unknown, callback: (err: Error | null, prime: unknown) => void): void;
export function generatePrime(size: number, options?: unknown, callback?: unknown): void {
    const prime = generatePrimeSync(size, options);
    if (typeof options === "function") {
        (options as (err: Error | null, prime: unknown) => void)(null, prime);
    } else if (typeof callback === "function") {
        (callback as (err: Error | null, prime: unknown) => void)(null, prime);
    }
}

export function generatePrimeSync(size: number, options?: unknown): unknown {
    if (!Number.isSafeInteger(size) || size < 2 || size > 52) {
        throw new RangeError("crypto.generatePrimeSync size must be between 2 and 52 bits");
    }
    const minimum = Math.pow(2, size - 1);
    let candidate = minimum + 1;
    if (candidate % 2 === 0) {
        candidate++;
    }
    while (!checkPrimeSync(candidate)) {
        candidate += 2;
    }
    return candidate;
}

export function getCiphers(): string[] {
    return ["aes-128-cbc", "aes-256-cbc", "aes-256-gcm"];
}

export function getCurves(): string[] {
    return ["prime256v1", "secp256k1"];
}

export function getHashes(): string[] {
    return ["sha256", "sha512", "md5"];
}

export function getRandomValues<T extends ArrayBufferView | null>(typedArray: T): T {
    if (typedArray === null || !ArrayBuffer.isView(typedArray)) {
        throw new TypeError("crypto.getRandomValues requires an ArrayBufferView");
    }
    const view = new Uint8Array(typedArray.buffer, typedArray.byteOffset, typedArray.byteLength);
    const random = randomBytes(view.length);
    for (let i = 0; i < view.length; i++) {
        view[i] = random[i];
    }
    return typedArray;
}

export function hkdf(digest: string, ikm: unknown, salt: unknown, info: unknown, keylen: number, callback: (err: Error | null, derivedKey: ArrayBuffer) => void): void {
    if (typeof ikm !== "string" || typeof salt !== "string" || typeof info !== "string") {
        throw new TypeError("crypto.hkdf currently requires string inputs");
    }
    callback(null, __scriptgo.hkdfSync(digest, ikm, salt, info, keylen));
}

export function hkdfSync(digest: string, ikm: unknown, salt: unknown, info: unknown, keylen: number): ArrayBuffer {
    if (typeof ikm !== "string" || typeof salt !== "string" || typeof info !== "string") {
        throw new TypeError("crypto.hkdfSync currently requires string inputs");
    }
    return __scriptgo.hkdfSync(digest, ikm, salt, info, keylen);
}

export function pbkdf2(password: unknown, salt: unknown, iterations: number, keylen: number, digest: string, callback: (err: Error | null, derivedKey: Buffer) => void): void {
    if (typeof password !== "string" || typeof salt !== "string") {
        throw new TypeError("crypto.pbkdf2 requires string password and salt");
    }
    callback(null, __scriptgo.pbkdf2Sync(password, salt, iterations, keylen, digest));
}

export function pbkdf2Sync(password: unknown, salt: unknown, iterations: number, keylen: number, digest: string): Buffer {
    if (typeof password !== "string" || typeof salt !== "string") {
        throw new TypeError("crypto.pbkdf2Sync requires string password and salt");
    }
    return __scriptgo.pbkdf2Sync(password, salt, iterations, keylen, digest);
}

export function randomBytes(size: number, callback?: (err: Error | null, buf: Buffer) => void): Buffer {
    const buf = __scriptgo.randomBytes(size);
    if (callback) {
        callback(null, buf);
    }
    return buf;
}

export type RandomFillCallback = (err: Error | null, buf: Buffer) => void;

export function randomFill(buffer: Buffer): void;
export function randomFill(buffer: Buffer, offset: number): void;
export function randomFill(buffer: Buffer, offset: number, size: number): void;
export function randomFill(buffer: Buffer, callback: RandomFillCallback): void;
export function randomFill(buffer: Buffer, offset: number, callback: RandomFillCallback): void;
export function randomFill(buffer: Buffer, offset: number, size: number, callback: RandomFillCallback): void;
export function randomFill(buffer: Buffer, offset?: number | RandomFillCallback, size?: number | RandomFillCallback, callback?: RandomFillCallback): void {
    if (!Buffer.isBuffer(buffer)) {
        throw new TypeError("crypto.randomFill requires a Buffer");
    }
    let filled: Buffer;
    if (typeof offset === "number" && typeof size === "number") {
        filled = __scriptgo.randomFill(buffer, offset, size);
    } else if (typeof offset === "number") {
        filled = __scriptgo.randomFill(buffer, offset);
    } else {
        filled = __scriptgo.randomFill(buffer);
    }
    if (typeof offset === "function") {
        offset(null, filled);
    } else if (typeof size === "function") {
        size(null, filled);
    } else if (typeof callback === "function") {
        callback(null, filled);
    }
}

export function randomFillSync(buffer: Buffer, offset?: number, size?: number): Buffer {
    return __scriptgo.randomFill(buffer, offset, size);
}

export function randomInt(min: number, max?: number, callback?: (err: Error | null, n: number) => void): number {
    const value = max === undefined ? __scriptgo.randomInt(0, min) : __scriptgo.randomInt(min, max);
    if (callback) callback(null, value);
    return value;
}

export function randomUUID(): string {
    return __scriptgo.randomUUID();
}

export function scrypt(password: unknown, salt: unknown, keylen: number, callback: (err: Error | null, derivedKey: Buffer) => void): void;
export function scrypt(password: unknown, salt: unknown, keylen: number, options: unknown, callback: (err: Error | null, derivedKey: Buffer) => void): void;
export function scrypt(password: unknown, salt: unknown, keylen: number, options?: unknown, callback?: unknown): void {
    if (typeof password !== "string" || typeof salt !== "string") {
        throw new TypeError("crypto.scrypt requires string password and salt");
    }
    const derivedKey = __scriptgo.scryptSync(password, salt, keylen);
    if (typeof options === "function") {
        (options as (err: Error | null, derivedKey: Buffer) => void)(null, derivedKey);
    } else if (typeof callback === "function") {
        (callback as (err: Error | null, derivedKey: Buffer) => void)(null, derivedKey);
    }
}

export function scryptSync(password: unknown, salt: unknown, keylen: number, options?: unknown): Buffer {
    if (typeof password !== "string" || typeof salt !== "string") {
        throw new TypeError("crypto.scryptSync requires string password and salt");
    }
    return __scriptgo.scryptSync(password, salt, keylen);
}

export function timingSafeEqual(a: Buffer, b: Buffer): boolean {
    return __scriptgo.timingSafeEqual(a, b);
}

export default {
    Hash,
    Hmac,
    X509Certificate,
    constants,
    subtle,
    webcrypto,
    checkPrime,
    checkPrimeSync,
    createHash,
    createHmac,
    generatePrime,
    generatePrimeSync,
    getCiphers,
    getCurves,
    getHashes,
    getRandomValues,
    hkdf,
    hkdfSync,
    pbkdf2,
    pbkdf2Sync,
    randomBytes,
    randomFill,
    randomFillSync,
    randomInt,
    randomUUID,
    scrypt,
    scryptSync,
    timingSafeEqual,
};
