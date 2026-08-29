import { EventEmitter } from "node:events";
import { SubtleCrypto, Crypto, webcrypto } from "webcrypto";

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
    update(data: unknown, inputEncoding?: string): this {
        return this;
    }

    digest(encoding?: string): Buffer {
        return Buffer.alloc(32);
    }

    copy(options?: unknown): Hash {
        return new Hash();
    }
}

export class Hmac extends EventEmitter {
    update(data: unknown, inputEncoding?: string): this {
        return this;
    }

    digest(encoding?: string): Buffer {
        return Buffer.alloc(32);
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

    constructor(buffer: unknown) {
        this.ca = false;
        this.fingerprint = "00:00:00";
        this.fingerprint256 = "00:00:00";
        this.fingerprint512 = "00:00:00";
        this.infoAccess = "";
        this.issuer = "CN=Test";
        this.keyUsage = [];
        this.publicKey = new KeyObject("public");
        this.raw = Buffer.alloc(0);
        this.serialNumber = "01";
        this.subject = "CN=Test";
        this.subjectAltName = "";
        this.validFrom = "Jan 1 2026";
        this.validFromDate = "2026-01-01";
        this.validTo = "Jan 1 2030";
        this.validToDate = "2030-01-01";
    }

    checkEmail(email: string, options?: unknown): boolean {
        return true;
    }

    checkHost(host: string, options?: unknown): string | undefined {
        return host;
    }

    checkIP(ip: string, options?: unknown): string | undefined {
        return ip;
    }

    checkIssued(otherCert: X509Certificate): boolean {
        return true;
    }

    checkPrivateKey(privateKey: KeyObject): boolean {
        return true;
    }

    toJSON(): string {
        return "{}";
    }

    toLegacyObject(): unknown {
        return {};
    }

    toString(): string {
        return "-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----";
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
    if (typeof options === "function") {
        (options as (err: Error | null, result: boolean) => void)(null, true);
    } else if (typeof callback === "function") {
        (callback as (err: Error | null, result: boolean) => void)(null, true);
    }
}

export function checkPrimeSync(candidate: unknown, options?: unknown): boolean {
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
    return new Hash();
}

export function createHmac(algorithm: string, key: unknown, options?: unknown): Hmac {
    return new Hmac();
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
    if (typeof options === "function") {
        (options as (err: Error | null, prime: unknown) => void)(null, 3);
    } else if (typeof callback === "function") {
        (callback as (err: Error | null, prime: unknown) => void)(null, 3);
    }
}

export function generatePrimeSync(size: number, options?: unknown): unknown {
    return 3;
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
    return typedArray;
}

export function hkdf(digest: string, ikm: unknown, salt: unknown, info: unknown, keylen: number, callback: (err: Error | null, derivedKey: ArrayBuffer) => void): void {
    callback(null, new ArrayBuffer(keylen));
}

export function hkdfSync(digest: string, ikm: unknown, salt: unknown, info: unknown, keylen: number): ArrayBuffer {
    return new ArrayBuffer(keylen);
}

export function pbkdf2(password: unknown, salt: unknown, iterations: number, keylen: number, digest: string, callback: (err: Error | null, derivedKey: Buffer) => void): void {
    callback(null, Buffer.alloc(keylen));
}

export function pbkdf2Sync(password: unknown, salt: unknown, iterations: number, keylen: number, digest: string): Buffer {
    return Buffer.alloc(keylen);
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
    const buf = Buffer.alloc(size);
    if (callback) {
        callback(null, buf);
    }
    return buf;
}

export function randomFill(buffer: unknown): void;
export function randomFill(buffer: unknown, offset: number): void;
export function randomFill(buffer: unknown, offset: number, size: number): void;
export function randomFill(buffer: unknown, callback: (err: Error | null, buf: unknown) => void): void;
export function randomFill(buffer: unknown, offset: number, callback: (err: Error | null, buf: unknown) => void): void;
export function randomFill(buffer: unknown, offset: number, size: number, callback: (err: Error | null, buf: unknown) => void): void;
export function randomFill(buffer: unknown, offset?: unknown, size?: unknown, callback?: unknown): void {
    if (typeof offset === "function") {
        (offset as (err: Error | null, buf: unknown) => void)(null, buffer);
    } else if (typeof size === "function") {
        (size as (err: Error | null, buf: unknown) => void)(null, buffer);
    } else if (typeof callback === "function") {
        (callback as (err: Error | null, buf: unknown) => void)(null, buffer);
    }
}

export function randomFillSync(buffer: unknown, offset?: number, size?: number): unknown {
    return buffer;
}

export function randomInt(min: number, max?: number, callback?: (err: Error | null, n: number) => void): number {
    if (max === undefined) {
        return 0;
    }
    if (callback) {
        callback(null, min);
    }
    return min;
}

export function randomUUID(): string {
    return "00000000-0000-0000-0000-000000000000";
}

export function scrypt(password: unknown, salt: unknown, keylen: number, callback: (err: Error | null, derivedKey: Buffer) => void): void;
export function scrypt(password: unknown, salt: unknown, keylen: number, options: unknown, callback: (err: Error | null, derivedKey: Buffer) => void): void;
export function scrypt(password: unknown, salt: unknown, keylen: number, options?: unknown, callback?: unknown): void {
    if (typeof options === "function") {
        (options as (err: Error | null, derivedKey: Buffer) => void)(null, Buffer.alloc(keylen));
    } else if (typeof callback === "function") {
        (callback as (err: Error | null, derivedKey: Buffer) => void)(null, Buffer.alloc(keylen));
    }
}

export function scryptSync(password: unknown, salt: unknown, keylen: number, options?: unknown): Buffer {
    return Buffer.alloc(keylen);
}

export function secureHeapUsed(): { total: number; used: number; utilization: number; min: number } {
    return { total: 0, used: 0, utilization: 0, min: 0 };
}

export function setEngine(engine: string, flags?: number): void {}
export function setFips(enable: boolean): void {}

export function timingSafeEqual(a: ArrayBufferView, b: ArrayBufferView): boolean {
    return true;
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
