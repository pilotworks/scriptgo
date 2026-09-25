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
    function cipherCreate(algorithm: string, key: Buffer, iv: Buffer, isEncrypt: number): object;
    function cipherUpdate(ctx: object, input: Buffer): Buffer;
    function cipherFinal(ctx: object): Buffer;
    function cipherSetAAD(ctx: object, aad: Buffer): void;
    function cipherGetTag(ctx: object): Buffer;
    function cipherSetTag(ctx: object, tag: Buffer): void;
    function cipherSetAutoPadding(ctx: object, autoPadding: number): void;
    function cipherDestroy(ctx: object): void;
    function getCipherInfo(nameOrNid: string): string;
    function cryptoSign(algorithm: string, data: Buffer, key: string): Buffer;
    function cryptoVerify(algorithm: string, data: Buffer, key: string, signature: Buffer): boolean;
    function publicEncrypt(key: string, buffer: Buffer, padding: number): Buffer;
    function publicDecrypt(key: string, buffer: Buffer, padding: number): Buffer;
    function privateEncrypt(key: string, buffer: Buffer, padding: number): Buffer;
    function privateDecrypt(key: string, buffer: Buffer, padding: number): Buffer;
    function generateKeyPairSync(type: string, modulusLength: number, namedCurve: string): string;
    function keyDetails(key: string): string;
    function dhCreate(prime: Buffer, generator: Buffer): object;
    function dhCreateGroup(name: string): object;
    function dhGenerateKeys(dh: object): Buffer;
    function dhComputeSecret(dh: object, otherPublicKey: Buffer): Buffer;
    function dhGetKey(dh: object, which: number): Buffer;
    function dhSetKey(dh: object, which: number, key: Buffer): void;
    function dhDestroy(dh: object): void;
    function ecdhCreate(curveName: string): object;
    function ecdhGenerateKeys(ecdh: object): Buffer;
    function ecdhComputeSecret(ecdh: object, otherPublicKey: Buffer): Buffer;
    function ecdhGetKey(ecdh: object, which: number): Buffer;
    function ecdhSetKey(ecdh: object, which: number, key: Buffer): void;
    function ecdhDestroy(ecdh: object): void;
    function spkacVerify(spkac: string): boolean;
    function spkacExportChallenge(spkac: string): string;
    function spkacExportPublicKey(spkac: string): string;
    function x509CheckPrivateKey(cert: string, key: string): boolean;
    function x509Verify(cert: string, pubkey: string): boolean;
    function getFips(): number;
    function setFips(enable: number): void;
    function secureHeapUsed(): string;
    function setEngine(engine: string, flags: number): void;
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

export type KeyObjectType = "secret" | "public" | "private";

export interface AsymmetricKeyDetails {
    modulusLength?: number;
    publicExponent?: bigint | number;
    hashAlgorithm?: string;
    mgf1HashAlgorithm?: string;
    saltLength?: number;
    divisorLength?: number;
    namedCurve?: string;
}

export class KeyObject {
    readonly type: KeyObjectType;
    readonly asymmetricKeyType?: string;
    readonly asymmetricKeyDetails?: AsymmetricKeyDetails;
    readonly symmetricKeySize?: number;
    private _rawKey: string | Buffer;

    constructor(type: KeyObjectType = "secret", rawKey: string | Buffer = "") {
        this.type = type;
        this._rawKey = rawKey;
        if (type === "secret") {
            this.symmetricKeySize = typeof rawKey === "string" ? Buffer.byteLength(rawKey) : (rawKey as Buffer).length;
            this.asymmetricKeyType = undefined;
            this.asymmetricKeyDetails = undefined;
        } else {
            const detailsRaw = __scriptgo.keyDetails(String(rawKey));
            try {
                const parsed = JSON.parse(detailsRaw) as { asymmetricKeyType?: string, modulusLength?: number };
                this.asymmetricKeyType = parsed.asymmetricKeyType || "rsa";
                this.asymmetricKeyDetails = { modulusLength: parsed.modulusLength || 2048 };
            } catch {
                this.asymmetricKeyType = "rsa";
                this.asymmetricKeyDetails = { modulusLength: 2048 };
            }
        }
    }

    export(options?: { type?: string, format?: string }): string | Buffer {
        if (this.type === "secret") {
            return typeof this._rawKey === "string" ? Buffer.from(this._rawKey) : this._rawKey;
        }
        return String(this._rawKey);
    }

    equals(other: KeyObject): boolean {
        if (!other || this.type !== other.type) return false;
        return String(this.export()) === String(other.export());
    }

    toCryptoKey(algorithm: unknown, extractable: boolean, keyUsages: string[]): unknown {
        return {
            type: this.type,
            extractable: extractable,
            algorithm: algorithm,
            usages: keyUsages
        };
    }
}

export function createSecretKey(key: CryptoBinary, encoding?: string): KeyObject {
    return new KeyObject("secret", toCryptoBuffer(key, encoding));
}

export function createPublicKey(key: string | Buffer | KeyObject): KeyObject {
    if (key instanceof KeyObject) return key;
    return new KeyObject("public", String(key));
}

export function createPrivateKey(key: string | Buffer | KeyObject): KeyObject {
    if (key instanceof KeyObject) return key;
    return new KeyObject("private", String(key));
}

export interface GenerateKeyPairOptions {
    modulusLength?: number;
    namedCurve?: string;
    publicKeyEncoding?: { type: string, format: string };
    privateKeyEncoding?: { type: string, format: string };
}

export function generateKeyPairSync(type: "rsa" | "ec" | "ed25519" | string, options: GenerateKeyPairOptions = {}): { publicKey: KeyObject | string, privateKey: KeyObject | string } {
    const modLen = options.modulusLength || 2048;
    const curve = options.namedCurve || "prime256v1";
    const raw = __scriptgo.generateKeyPairSync(type, modLen, curve);
    const parsed = JSON.parse(raw) as { publicKey: string, privateKey: string };
    if (options.publicKeyEncoding) {
        return { publicKey: parsed.publicKey, privateKey: parsed.privateKey };
    }
    return {
        publicKey: new KeyObject("public", parsed.publicKey),
        privateKey: new KeyObject("private", parsed.privateKey)
    };
}

export function generateKeyPair(type: string, options: GenerateKeyPairOptions, callback: (err: Error | null, publicKey: unknown, privateKey: unknown) => void): void {
    try {
        const res = generateKeyPairSync(type, options);
        callback(null, res.publicKey, res.privateKey);
    } catch (e: unknown) {
        callback(e as Error, null, null);
    }
}

export function generateKeySync(type: "hmac" | "aes", options: { length: number }): KeyObject {
    const len = options.length ? Math.floor(options.length / 8) : 32;
    const bytes = __scriptgo.randomBytes(len);
    return new KeyObject("secret", bytes);
}

export function generateKey(type: "hmac" | "aes", options: { length: number }, callback: (err: Error | null, key: KeyObject | null) => void): void {
    try {
        const key = generateKeySync(type, options);
        callback(null, key);
    } catch (e: unknown) {
        callback(e as Error, null);
    }
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
    publicKey: KeyObject = new KeyObject("public", "");
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

    checkPrivateKey(privateKey: KeyObject): boolean {
        const keyPem = String(privateKey.export());
        return __scriptgo.x509CheckPrivateKey(this._pem, keyPem);
    }

    verify(publicKey: KeyObject): boolean {
        const pubPem = String(publicKey.export());
        return __scriptgo.x509Verify(this._pem, pubPem);
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

export class Certificate {
    constructor() {}

    static verifySpkac(spkac: CryptoBinary, encoding?: string): boolean {
        const str = typeof spkac === "string" ? spkac : toCryptoBuffer(spkac, encoding).toString("utf8");
        return __scriptgo.spkacVerify(str);
    }

    static exportChallenge(spkac: CryptoBinary, encoding?: string): Buffer {
        const str = typeof spkac === "string" ? spkac : toCryptoBuffer(spkac, encoding).toString("utf8");
        const chal = __scriptgo.spkacExportChallenge(str);
        return Buffer.from(chal, "utf8");
    }

    static exportPublicKey(spkac: CryptoBinary, encoding?: string): Buffer {
        const str = typeof spkac === "string" ? spkac : toCryptoBuffer(spkac, encoding).toString("utf8");
        const pub = __scriptgo.spkacExportPublicKey(str);
        return Buffer.from(pub, "utf8");
    }

    verifySpkac(spkac: CryptoBinary, encoding?: string): boolean {
        return Certificate.verifySpkac(spkac, encoding);
    }

    exportChallenge(spkac: CryptoBinary, encoding?: string): Buffer {
        return Certificate.exportChallenge(spkac, encoding);
    }

    exportPublicKey(spkac: CryptoBinary, encoding?: string): Buffer {
        return Certificate.exportPublicKey(spkac, encoding);
    }
}

export class Cipher extends EventEmitter {
    private _ctx: object;

    constructor(algorithm: string, key: CryptoBinary | KeyObject, iv: CryptoBinary | null, options?: unknown) {
        super();
        const keyBuf = key instanceof KeyObject ? toCryptoBuffer(key.export()) : toCryptoBuffer(key);
        const ivBuf = iv ? toCryptoBuffer(iv) : Buffer.alloc(0);
        this._ctx = __scriptgo.cipherCreate(algorithm, keyBuf, ivBuf, 1);
    }

    update(data: CryptoBinary, inputEncoding?: string): Buffer;
    update(data: CryptoBinary, inputEncoding: string | undefined, outputEncoding: string): string;
    update(data: CryptoBinary, inputEncoding?: string, outputEncoding?: string): Buffer | string;
    update(data: CryptoBinary, inputEncoding?: string, outputEncoding?: string): Buffer | string {
        const inBuf = toCryptoBuffer(data, inputEncoding);
        const outBuf = __scriptgo.cipherUpdate(this._ctx, inBuf);
        if (outputEncoding !== undefined) {
            return outBuf.toString(outputEncoding);
        }
        return outBuf;
    }

    final(): Buffer;
    final(outputEncoding: string): string;
    final(outputEncoding?: string): Buffer | string;
    final(outputEncoding?: string): Buffer | string {
        const outBuf = __scriptgo.cipherFinal(this._ctx);
        if (outputEncoding !== undefined) {
            return outBuf.toString(outputEncoding);
        }
        return outBuf;
    }

    setAAD(buffer: Buffer, options?: unknown): this {
        __scriptgo.cipherSetAAD(this._ctx, buffer);
        return this;
    }

    getAuthTag(): Buffer {
        return __scriptgo.cipherGetTag(this._ctx);
    }

    setAutoPadding(autoPadding: boolean = true): this {
        __scriptgo.cipherSetAutoPadding(this._ctx, autoPadding ? 1 : 0);
        return this;
    }
}

export class Decipher extends EventEmitter {
    private _ctx: object;

    constructor(algorithm: string, key: CryptoBinary | KeyObject, iv: CryptoBinary | null, options?: unknown) {
        super();
        const keyBuf = key instanceof KeyObject ? toCryptoBuffer(key.export()) : toCryptoBuffer(key);
        const ivBuf = iv ? toCryptoBuffer(iv) : Buffer.alloc(0);
        this._ctx = __scriptgo.cipherCreate(algorithm, keyBuf, ivBuf, 0);
    }

    update(data: CryptoBinary, inputEncoding?: string): Buffer;
    update(data: CryptoBinary, inputEncoding: string | undefined, outputEncoding: string): string;
    update(data: CryptoBinary, inputEncoding?: string, outputEncoding?: string): Buffer | string;
    update(data: CryptoBinary, inputEncoding?: string, outputEncoding?: string): Buffer | string {
        const inBuf = toCryptoBuffer(data, inputEncoding);
        const outBuf = __scriptgo.cipherUpdate(this._ctx, inBuf);
        if (outputEncoding !== undefined) {
            return outBuf.toString(outputEncoding);
        }
        return outBuf;
    }

    final(): Buffer;
    final(outputEncoding: string): string;
    final(outputEncoding?: string): Buffer | string;
    final(outputEncoding?: string): Buffer | string {
        const outBuf = __scriptgo.cipherFinal(this._ctx);
        if (outputEncoding !== undefined) {
            return outBuf.toString(outputEncoding);
        }
        return outBuf;
    }

    setAAD(buffer: Buffer, options?: unknown): this {
        __scriptgo.cipherSetAAD(this._ctx, buffer);
        return this;
    }

    setAuthTag(buffer: Buffer, encoding?: string): this {
        __scriptgo.cipherSetTag(this._ctx, buffer);
        return this;
    }

    setAutoPadding(autoPadding: boolean = true): this {
        __scriptgo.cipherSetAutoPadding(this._ctx, autoPadding ? 1 : 0);
        return this;
    }
}

export function createCipheriv(algorithm: string, key: CryptoBinary | KeyObject, iv: CryptoBinary | null, options?: unknown): Cipher {
    return new Cipher(algorithm, key, iv, options);
}

export function createDecipheriv(algorithm: string, key: CryptoBinary | KeyObject, iv: CryptoBinary | null, options?: unknown): Decipher {
    return new Decipher(algorithm, key, iv, options);
}

export interface CipherInfo {
    name: string;
    nid: number;
    blockSize: number;
    ivLength: number;
    keyLength: number;
    mode: string;
}

export function getCipherInfo(nameOrNid: string | number, options?: unknown): CipherInfo | undefined {
    const raw = __scriptgo.getCipherInfo(String(nameOrNid));
    if (!raw || raw === "{}") return undefined;
    try {
        return JSON.parse(raw) as CipherInfo;
    } catch {
        return undefined;
    }
}

export class Sign extends EventEmitter {
    private _algorithm: string;
    private _data: Buffer;

    constructor(algorithm: string = "sha256", options?: unknown) {
        super();
        this._algorithm = algorithm;
        this._data = Buffer.alloc(0);
    }

    update(data: CryptoBinary, inputEncoding?: string): this {
        this._data = Buffer.concat([this._data, toCryptoBuffer(data, inputEncoding)]);
        return this;
    }

    sign(privateKey: KeyObject | string | { key: string | KeyObject, passphrase?: string }): Buffer;
    sign(privateKey: KeyObject | string | { key: string | KeyObject, passphrase?: string }, outputEncoding: string): string;
    sign(privateKey: KeyObject | string | { key: string | KeyObject, passphrase?: string }, outputEncoding?: string): Buffer | string;
    sign(privateKey: KeyObject | string | { key: string | KeyObject, passphrase?: string }, outputEncoding?: string): Buffer | string {
        let keyPem = "";
        if (privateKey instanceof KeyObject) {
            keyPem = String(privateKey.export());
        } else if (typeof privateKey === "string") {
            keyPem = privateKey;
        } else if (typeof privateKey === "object" && privateKey !== null) {
            const k = (privateKey as { key: string | KeyObject }).key;
            keyPem = k instanceof KeyObject ? String(k.export()) : String(k);
        }
        const sigBuf = __scriptgo.cryptoSign(this._algorithm, this._data, keyPem);
        if (outputEncoding !== undefined) {
            return sigBuf.toString(outputEncoding);
        }
        return sigBuf;
    }
}

export class Verify extends EventEmitter {
    private _algorithm: string;
    private _data: Buffer;

    constructor(algorithm: string = "sha256", options?: unknown) {
        super();
        this._algorithm = algorithm;
        this._data = Buffer.alloc(0);
    }

    update(data: CryptoBinary, inputEncoding?: string): this {
        this._data = Buffer.concat([this._data, toCryptoBuffer(data, inputEncoding)]);
        return this;
    }

    verify(publicKey: KeyObject | string | { key: string | KeyObject }, signature: CryptoBinary, signatureEncoding?: string): boolean {
        let keyPem = "";
        if (publicKey instanceof KeyObject) {
            keyPem = String(publicKey.export());
        } else if (typeof publicKey === "string") {
            keyPem = publicKey;
        } else if (typeof publicKey === "object" && publicKey !== null) {
            const k = (publicKey as { key: string | KeyObject }).key;
            keyPem = k instanceof KeyObject ? String(k.export()) : String(k);
        }
        const sigBuf = toCryptoBuffer(signature, signatureEncoding);
        return __scriptgo.cryptoVerify(this._algorithm, this._data, keyPem, sigBuf);
    }
}

export function createSign(algorithm: string, options?: unknown): Sign {
    return new Sign(algorithm, options);
}

export function createVerify(algorithm: string, options?: unknown): Verify {
    return new Verify(algorithm, options);
}

export class DiffieHellman {
    protected _handle: object;
    verifyError: number = 0;

    constructor(prime: CryptoBinary | number, primeEncoding?: string | number, generator?: CryptoBinary | number, generatorEncoding?: string) {
        let primeBuf: Buffer;
        let genBuf: Buffer = Buffer.alloc(0);
        if (typeof prime === "number") {
            const pGroup = "FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7EDEE386BFB5A899FA5AE9F24117C4B1FE649286651ECE65381FFFFFFFFFFFFFFFF";
            primeBuf = Buffer.from(pGroup, "hex");
            genBuf = Buffer.from([typeof primeEncoding === "number" ? primeEncoding : 2]);
        } else {
            primeBuf = toCryptoBuffer(prime, typeof primeEncoding === "string" ? primeEncoding : undefined);
            if (generator !== undefined) {
                if (typeof generator === "number") {
                    genBuf = Buffer.from([generator]);
                } else {
                    genBuf = toCryptoBuffer(generator, generatorEncoding);
                }
            } else {
                genBuf = Buffer.from([2]);
            }
        }
        this._handle = __scriptgo.dhCreate(primeBuf, genBuf);
    }

    generateKeys(): Buffer;
    generateKeys(encoding: string): string;
    generateKeys(encoding?: string): Buffer | string;
    generateKeys(encoding?: string): Buffer | string {
        const buf = __scriptgo.dhGenerateKeys(this._handle);
        return encoding !== undefined ? buf.toString(encoding) : buf;
    }

    computeSecret(otherPublicKey: CryptoBinary): Buffer;
    computeSecret(otherPublicKey: CryptoBinary, outputEncoding: string): string;
    computeSecret(otherPublicKey: CryptoBinary, inputEncoding: string, outputEncoding: string): string;
    computeSecret(otherPublicKey: CryptoBinary, inputEncoding?: string, outputEncoding?: string): Buffer | string;
    computeSecret(otherPublicKey: CryptoBinary, inputEncoding?: string, outputEncoding?: string): Buffer | string {
        const otherBuf = toCryptoBuffer(otherPublicKey, inputEncoding);
        const sec = __scriptgo.dhComputeSecret(this._handle, otherBuf);
        return outputEncoding !== undefined ? sec.toString(outputEncoding) : sec;
    }

    getPrime(): Buffer;
    getPrime(encoding: string): string;
    getPrime(encoding?: string): Buffer | string;
    getPrime(encoding?: string): Buffer | string {
        const buf = __scriptgo.dhGetKey(this._handle, 0);
        return encoding !== undefined ? buf.toString(encoding) : buf;
    }

    getGenerator(): Buffer;
    getGenerator(encoding: string): string;
    getGenerator(encoding?: string): Buffer | string;
    getGenerator(encoding?: string): Buffer | string {
        const buf = __scriptgo.dhGetKey(this._handle, 1);
        return encoding !== undefined ? buf.toString(encoding) : buf;
    }

    getPublicKey(): Buffer;
    getPublicKey(encoding: string): string;
    getPublicKey(encoding?: string): Buffer | string;
    getPublicKey(encoding?: string): Buffer | string {
        const buf = __scriptgo.dhGetKey(this._handle, 2);
        return encoding !== undefined ? buf.toString(encoding) : buf;
    }

    getPrivateKey(): Buffer;
    getPrivateKey(encoding: string): string;
    getPrivateKey(encoding?: string): Buffer | string;
    getPrivateKey(encoding?: string): Buffer | string {
        const buf = __scriptgo.dhGetKey(this._handle, 3);
        return encoding !== undefined ? buf.toString(encoding) : buf;
    }

    setPublicKey(publicKey: CryptoBinary, encoding?: string): this {
        __scriptgo.dhSetKey(this._handle, 2, toCryptoBuffer(publicKey, encoding));
        return this;
    }

    setPrivateKey(privateKey: CryptoBinary, encoding?: string): this {
        __scriptgo.dhSetKey(this._handle, 3, toCryptoBuffer(privateKey, encoding));
        return this;
    }
}

export class DiffieHellmanGroup extends DiffieHellman {
    constructor(name: string) {
        super(2048);
        this._handle = __scriptgo.dhCreateGroup(name);
    }
}

export function createDiffieHellman(prime: unknown, primeEncoding?: unknown, generator?: unknown, generatorEncoding?: unknown): DiffieHellman {
    return new DiffieHellman(prime as any, primeEncoding as any, generator as any, generatorEncoding as any);
}

export function createDiffieHellmanGroup(name: string): DiffieHellmanGroup {
    return new DiffieHellmanGroup(name);
}

export function getDiffieHellman(name: string): DiffieHellmanGroup {
    return new DiffieHellmanGroup(name);
}

export class ECDH {
    private _handle: object;

    constructor(curveName: string) {
        this._handle = __scriptgo.ecdhCreate(curveName);
    }

    generateKeys(): Buffer;
    generateKeys(encoding: string, format?: string): string;
    generateKeys(encoding?: string, format?: string): Buffer | string;
    generateKeys(encoding?: string, format?: string): Buffer | string {
        const buf = __scriptgo.ecdhGenerateKeys(this._handle);
        return encoding !== undefined ? buf.toString(encoding) : buf;
    }

    computeSecret(otherPublicKey: CryptoBinary): Buffer;
    computeSecret(otherPublicKey: CryptoBinary, outputEncoding: string): string;
    computeSecret(otherPublicKey: CryptoBinary, inputEncoding: string, outputEncoding: string): string;
    computeSecret(otherPublicKey: CryptoBinary, inputEncoding?: string, outputEncoding?: string): Buffer | string;
    computeSecret(otherPublicKey: CryptoBinary, inputEncoding?: string, outputEncoding?: string): Buffer | string {
        const otherBuf = toCryptoBuffer(otherPublicKey, inputEncoding);
        const sec = __scriptgo.ecdhComputeSecret(this._handle, otherBuf);
        return outputEncoding !== undefined ? sec.toString(outputEncoding) : sec;
    }

    getPublicKey(): Buffer;
    getPublicKey(encoding: string, format?: string): string;
    getPublicKey(encoding?: string, format?: string): Buffer | string;
    getPublicKey(encoding?: string, format?: string): Buffer | string {
        const buf = __scriptgo.ecdhGetKey(this._handle, 0);
        return encoding !== undefined ? buf.toString(encoding) : buf;
    }

    getPrivateKey(): Buffer;
    getPrivateKey(encoding: string): string;
    getPrivateKey(encoding?: string): Buffer | string;
    getPrivateKey(encoding?: string): Buffer | string {
        const buf = __scriptgo.ecdhGetKey(this._handle, 1);
        return encoding !== undefined ? buf.toString(encoding) : buf;
    }

    setPublicKey(publicKey: CryptoBinary, encoding?: string): this {
        __scriptgo.ecdhSetKey(this._handle, 0, toCryptoBuffer(publicKey, encoding));
        return this;
    }

    setPrivateKey(privateKey: CryptoBinary, encoding?: string): this {
        __scriptgo.ecdhSetKey(this._handle, 1, toCryptoBuffer(privateKey, encoding));
        return this;
    }
}

export function createECDH(curveName: string): ECDH {
    return new ECDH(curveName);
}

function extractKeyPemAndPadding(key: KeyObject | string | { key: string | KeyObject, padding?: number }): { pem: string, padding: number } {
    let pem = "";
    let padding = 1;
    if (key instanceof KeyObject) {
        pem = String(key.export());
    } else if (typeof key === "string") {
        pem = key;
    } else if (typeof key === "object" && key !== null) {
        const k = (key as { key: string | KeyObject, padding?: number }).key;
        pem = k instanceof KeyObject ? String(k.export()) : String(k);
        if (typeof (key as { padding?: number }).padding === "number") {
            padding = (key as { padding?: number }).padding!;
        }
    }
    return { pem, padding };
}

export function publicEncrypt(key: KeyObject | string | { key: string | KeyObject, padding?: number }, buffer: CryptoBinary): Buffer {
    const { pem, padding } = extractKeyPemAndPadding(key);
    return __scriptgo.publicEncrypt(pem, toCryptoBuffer(buffer), padding);
}

export function publicDecrypt(key: KeyObject | string | { key: string | KeyObject, padding?: number }, buffer: CryptoBinary): Buffer {
    const { pem, padding } = extractKeyPemAndPadding(key);
    return __scriptgo.publicDecrypt(pem, toCryptoBuffer(buffer), padding);
}

export function privateEncrypt(key: KeyObject | string | { key: string | KeyObject, padding?: number }, buffer: CryptoBinary): Buffer {
    const { pem, padding } = extractKeyPemAndPadding(key);
    return __scriptgo.privateEncrypt(pem, toCryptoBuffer(buffer), padding);
}

export function privateDecrypt(key: KeyObject | string | { key: string | KeyObject, padding?: number }, buffer: CryptoBinary): Buffer {
    const { pem, padding } = extractKeyPemAndPadding(key);
    return __scriptgo.privateDecrypt(pem, toCryptoBuffer(buffer), padding);
}

export function getFips(): number {
    return __scriptgo.getFips();
}

export function setFips(enable: boolean | number): void {
    __scriptgo.setFips(enable ? 1 : 0);
}

export function secureHeapUsed(): unknown {
    const raw = __scriptgo.secureHeapUsed();
    try {
        return JSON.parse(raw);
    } catch {
        return { total: 0, min: 0, used: 0, util: 0 };
    }
}

export function setEngine(engine: string, flags?: number): void {
    __scriptgo.setEngine(engine, flags || 0);
}

export let fips: number = 0;

export default {
    Certificate,
    Cipher,
    Decipher,
    DiffieHellman,
    DiffieHellmanGroup,
    ECDH,
    Hash,
    Hmac,
    KeyObject,
    Sign,
    Verify,
    X509Certificate,
    constants,
    subtle,
    webcrypto,
    checkPrime,
    checkPrimeSync,
    createCipheriv,
    createDecipheriv,
    createDiffieHellman,
    createDiffieHellmanGroup,
    createECDH,
    createHash,
    createHmac,
    createPrivateKey,
    createPublicKey,
    createSecretKey,
    createSign,
    createVerify,
    fips,
    generateKey,
    generateKeyPair,
    generateKeyPairSync,
    generateKeySync,
    generatePrime,
    generatePrimeSync,
    getCipherInfo,
    getCiphers,
    getCurves,
    getDiffieHellman,
    getFips,
    getHashes,
    getRandomValues,
    hkdf,
    hkdfSync,
    pbkdf2,
    pbkdf2Sync,
    privateDecrypt,
    privateEncrypt,
    publicDecrypt,
    publicEncrypt,
    randomBytes,
    randomFill,
    randomFillSync,
    randomInt,
    randomUUID,
    scrypt,
    scryptSync,
    secureHeapUsed,
    setEngine,
    setFips,
    timingSafeEqual,
};
