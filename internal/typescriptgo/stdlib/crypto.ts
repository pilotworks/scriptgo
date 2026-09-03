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

export class Certificate {
    static exportChallenge(spkac: unknown): Buffer {
        return Buffer.alloc(0);
    }

    static exportPublicKey(spkac: unknown): Buffer {
        return Buffer.alloc(0);
    }

    static verifySpkac(spkac: unknown): boolean {
        return true;
    }
}

export class Cipher extends EventEmitter {
    update(data: unknown, inputEncoding?: string, outputEncoding?: string): Buffer {
        return Buffer.alloc(0);
    }

    final(outputEncoding?: string): Buffer {
        return Buffer.alloc(0);
    }

    setAutoPadding(autoPadding?: boolean): this {
        return this;
    }

    getAuthTag(): Buffer {
        return Buffer.alloc(16);
    }

    setAAD(buffer: unknown, options?: unknown): this {
        return this;
    }
}

export class Decipher extends EventEmitter {
    update(data: unknown, inputEncoding?: string, outputEncoding?: string): Buffer {
        return Buffer.alloc(0);
    }

    final(outputEncoding?: string): Buffer {
        return Buffer.alloc(0);
    }

    setAutoPadding(autoPadding?: boolean): this {
        return this;
    }

    setAuthTag(buffer: unknown, encoding?: string): this {
        return this;
    }

    setAAD(buffer: unknown, options?: unknown): this {
        return this;
    }
}

export class DiffieHellman {
    verifyError: number = 0;

    generateKeys(encoding?: string): Buffer {
        return Buffer.alloc(32);
    }

    computeSecret(otherPublicKey: unknown, inputEncoding?: string, outputEncoding?: string): Buffer {
        return Buffer.alloc(32);
    }

    getPrime(encoding?: string): Buffer {
        return Buffer.alloc(32);
    }

    getGenerator(encoding?: string): Buffer {
        return Buffer.alloc(1);
    }

    getPublicKey(encoding?: string): Buffer {
        return Buffer.alloc(32);
    }

    getPrivateKey(encoding?: string): Buffer {
        return Buffer.alloc(32);
    }

    setPublicKey(publicKey: unknown, encoding?: string): this {
        return this;
    }

    setPrivateKey(privateKey: unknown, encoding?: string): this {
        return this;
    }
}

export class DiffieHellmanGroup extends DiffieHellman {}

export interface HashOptions {
    outputLength?: number;
}

export class ECDH {
    generateKeys(encoding?: string, format?: string): Buffer {
        return Buffer.alloc(32);
    }

    computeSecret(otherPublicKey: unknown, inputEncoding?: string, outputEncoding?: string): Buffer {
        return Buffer.alloc(32);
    }

    getPublicKey(encoding?: string, format?: string): Buffer {
        return Buffer.alloc(32);
    }

    getPrivateKey(encoding?: string): Buffer {
        return Buffer.alloc(32);
    }

    setPublicKey(publicKey: unknown, encoding?: string): this {
        return this;
    }

    setPrivateKey(privateKey: unknown, encoding?: string): this {
        return this;
    }
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

export class KeyObject {
    type: string = "secret";
    asymmetricKeyType: string = "";
    symmetricKeySize: number = 32;

    constructor(type: string = "secret") {
        this.type = type;
    }

    asymmetricKeyDetails(): unknown {
        return {};
    }

    equals(otherKeyObject: KeyObject): boolean {
        return true;
    }

    export(options?: unknown): unknown {
        return Buffer.alloc(0);
    }

    toCryptoKey(algorithm: unknown, extractable: boolean, keyUsages: string[]): unknown {
        return {};
    }
}

export class Sign extends EventEmitter {
    update(data: unknown, inputEncoding?: string): this {
        return this;
    }

    sign(privateKey: unknown, outputEncoding?: string): Buffer {
        return Buffer.alloc(64);
    }
}

export class Verify extends EventEmitter {
    update(data: unknown, inputEncoding?: string): this {
        return this;
    }

    verify(object: unknown, signature: unknown, signatureEncoding?: string): boolean {
        return true;
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
    publicKey: KeyObject = new KeyObject("public");
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
                b64 = certStr.substring(beginIdx + 27, endIdx).trim();
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
        if (this.subject === "" && certStr.length > 0) {
            this.subject = "CN=" + certStr.substring(0, 32);
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

    checkPrivateKey(privateKey: KeyObject): boolean {
        return true;
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

    verify(publicKey: KeyObject): boolean {
        return true;
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

export const fips: boolean = false;
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

export function createCipheriv(algorithm: string, key: unknown, iv: unknown, options?: unknown): Cipher {
    return new Cipher();
}

export function createDecipheriv(algorithm: string, key: unknown, iv: unknown, options?: unknown): Decipher {
    return new Decipher();
}

export function createDiffieHellman(primeLength: number, generator?: unknown): DiffieHellman {
    return new DiffieHellman();
}

export function createDiffieHellmanGroup(name: string): DiffieHellmanGroup {
    return new DiffieHellmanGroup();
}

export function createECDH(curveName: string): ECDH {
    return new ECDH();
}

export function createHash(algorithm: string, options?: unknown): Hash {
    return new Hash(algorithm);
}

export function createHmac(algorithm: string, key: CryptoBinary, options?: unknown): Hmac {
    return new Hmac(algorithm, key);
}

export function createPrivateKey(key: unknown): KeyObject {
    return new KeyObject("private");
}

export function createPublicKey(key: unknown): KeyObject {
    return new KeyObject("public");
}

export function createSecretKey(key: unknown, encoding?: string): KeyObject {
    return new KeyObject("secret");
}

export function createSign(algorithm: string, options?: unknown): Sign {
    return new Sign();
}

export function createVerify(algorithm: string, options?: unknown): Verify {
    return new Verify();
}

export function generateKey(type: string, options: unknown, callback: (err: Error | null, key: KeyObject) => void): void {
    callback(null, new KeyObject("secret"));
}

export function generateKeySync(type: string, options: unknown): KeyObject {
    return new KeyObject("secret");
}

export function generateKeyPair(type: string, options: unknown, callback: (err: Error | null, publicKey: KeyObject, privateKey: KeyObject) => void): void {
    callback(null, new KeyObject("public"), new KeyObject("private"));
}

export function generateKeyPairSync(type: string, options: unknown): { publicKey: KeyObject; privateKey: KeyObject } {
    return { publicKey: new KeyObject("public"), privateKey: new KeyObject("private") };
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

export function getCipherInfo(nameOrNid: string | number, options?: unknown): unknown {
    return { name: "aes-256-gcm", nid: 1, blockSize: 16, ivLength: 12, keyLength: 32 };
}

export function getCiphers(): string[] {
    return ["aes-128-cbc", "aes-256-cbc", "aes-256-gcm"];
}

export function getCurves(): string[] {
    return ["prime256v1", "secp256k1"];
}

export function getDiffieHellman(groupName: string): DiffieHellmanGroup {
    return new DiffieHellmanGroup();
}

export function getFips(): boolean {
    return false;
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

export function privateDecrypt(privateKey: unknown, buffer: unknown): Buffer {
    return Buffer.alloc(0);
}

export function privateEncrypt(privateKey: unknown, buffer: unknown): Buffer {
    return Buffer.alloc(0);
}

export function publicDecrypt(publicKey: unknown, buffer: unknown): Buffer {
    return Buffer.alloc(0);
}

export function publicEncrypt(publicKey: unknown, buffer: unknown): Buffer {
    return Buffer.alloc(0);
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

export function secureHeapUsed(): { total: number; used: number; utilization: number; min: number } {
    return { total: 0, used: 0, utilization: 0, min: 0 };
}

export function setEngine(engine: string, flags?: number): void {}
export function setFips(enable: boolean): void {}

export function timingSafeEqual(a: Buffer, b: Buffer): boolean {
    return __scriptgo.timingSafeEqual(a, b);
}

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
    fips,
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
    generateKey,
    generateKeySync,
    generateKeyPair,
    generateKeyPairSync,
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
