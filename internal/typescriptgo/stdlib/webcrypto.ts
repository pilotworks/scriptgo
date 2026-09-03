// Node.js Web Crypto API module (node:crypto.webcrypto or webcrypto)

declare namespace __scriptgo {
    function hashDigest(algorithm: string, data: string, encoding?: string): string;
    function hmacDigest(algorithm: string, key: string, data: string, encoding?: string): string;
    function pbkdf2Sync(password: string, salt: string, iterations: number, keylen: number, digest?: string): Buffer;
    function randomUUID(): string;
    function randomBytes(size: number): Buffer;
}

function sourceBytes(source: BufferSource): Buffer {
    let view: Uint8Array;
    if (source instanceof ArrayBuffer) {
        view = new Uint8Array(source);
    } else {
        view = source as Uint8Array;
    }
    const result = Buffer.alloc(view.length);
    for (let i = 0; i < view.length; i++) {
        result[i] = view[i];
    }
    return result;
}

function binaryString(bytes: Buffer): string {
    let result = "";
    for (let i = 0; i < bytes.length; i++) {
        result = result + String.fromCharCode(bytes[i]);
    }
    return result;
}

export class CryptoKey {
    type: string = "secret";
    extractable: boolean = true;
    algorithm: KeyAlgorithm = new KeyAlgorithm();
    usages: string[] = ["encrypt", "decrypt"];
    material: Buffer = Buffer.alloc(0);

    constructor(type: string = "secret", extractable: boolean = true, algorithm: KeyAlgorithm = new KeyAlgorithm(), usages: string[] = ["encrypt", "decrypt"]) {
        this.type = type;
        this.extractable = extractable;
        this.algorithm = algorithm;
        this.usages = usages;
    }
}

export class CryptoKeyPair {
    privateKey: CryptoKey = new CryptoKey("private");
    publicKey: CryptoKey = new CryptoKey("public");

    constructor(privateKey?: CryptoKey, publicKey?: CryptoKey) {
        this.privateKey = privateKey || new CryptoKey("private");
        this.publicKey = publicKey || new CryptoKey("public");
    }
}

export class Algorithm {
    name: string = "";
    constructor(name: string = "") {
        this.name = name;
    }
}

export class KeyAlgorithm {
    name: string = "AES-GCM";
    constructor(name: string = "AES-GCM") {
        this.name = name;
    }
}

export class AesDerivedKeyParams {
    name: string = "AES-CTR";
    length: number = 256;
}

export class AesCbcParams {
    name: string = "AES-CBC";
    iv: BufferSource = new Uint8Array(16);
}

export class AesCtrParams {
    name: string = "AES-CTR";
    counter: BufferSource = new Uint8Array(16);
    length: number = 64;
}

export class AesGcmParams {
    name: string = "AES-GCM";
    iv: BufferSource = new Uint8Array(12);
    additionalData: BufferSource = new Uint8Array(0);
    tagLength: number = 128;
}

export class AesKeyAlgorithm {
    name: string = "AES-GCM";
    length: number = 256;
}

export class AesKeyGenParams {
    name: string = "AES-GCM";
    length: number = 256;
}

export class EcdhKeyDeriveParams {
    name: string = "ECDH";
    public: CryptoKey = new CryptoKey("public");

    constructor() {
        this.name = "ECDH";
        this.public = new CryptoKey("public");
    }
}

export class EcdsaParams {
    name: string = "ECDSA";
    hash: string = "SHA-256";
}

export class EcKeyAlgorithm {
    name: string = "ECDSA";
    namedCurve: string = "P-256";
}

export class EcKeyGenParams {
    name: string = "ECDSA";
    namedCurve: string = "P-256";
}

export class EcKeyImportParams {
    name: string = "ECDSA";
    namedCurve: string = "P-256";
}

export class Ed448Params {
    name: string = "Ed448";
    context: BufferSource = new Uint8Array(0);
}

export class HkdfParams {
    name: string = "HKDF";
    hash: string = "SHA-256";
    salt: BufferSource = new Uint8Array(0);
    info: BufferSource = new Uint8Array(0);
}

export class HmacImportParams {
    name: string = "HMAC";
    hash: string = "SHA-256";
    length: number = 256;
}

export class HmacKeyAlgorithm {
    name: string = "HMAC";
    hash: KeyAlgorithm = new KeyAlgorithm("SHA-256");
    length: number = 256;

    constructor() {
        this.name = "HMAC";
        this.hash = new KeyAlgorithm("SHA-256");
        this.length = 256;
    }
}

export class HmacKeyGenParams {
    name: string = "HMAC";
    hash: string = "SHA-256";
    length: number = 256;
}

export class Pbkdf2Params {
    name: string = "PBKDF2";
    hash: string = "SHA-256";
    salt: BufferSource = new Uint8Array(0);
    iterations: number = 100000;
}

export class RsaHashedImportParams {
    name: string = "RSA-OAEP";
    hash: string = "SHA-256";
}

export class RsaHashedKeyAlgorithm {
    name: string = "RSA-OAEP";
    modulusLength: number = 2048;
    publicExponent: Uint8Array = new Uint8Array([1, 0, 1]);
    hash: KeyAlgorithm = new KeyAlgorithm("SHA-256");

    constructor() {
        this.name = "RSA-OAEP";
        this.modulusLength = 2048;
        this.publicExponent = new Uint8Array([1, 0, 1]);
        this.hash = new KeyAlgorithm("SHA-256");
    }
}

export class RsaHashedKeyGenParams {
    name: string = "RSA-OAEP";
    modulusLength: number = 2048;
    publicExponent: Uint8Array = new Uint8Array([1, 0, 1]);
    hash: string = "SHA-256";
}

export class RsaOaepParams {
    name: string = "RSA-OAEP";
    label: BufferSource = new Uint8Array(0);
}

export class RsaPssParams {
    name: string = "RSA-PSS";
    saltLength: number = 32;
}

export class SubtleCrypto {
    async sign(algorithm: unknown, key: CryptoKey, data: BufferSource): Promise<ArrayBuffer> {
        let name = key.algorithm.name;
        if (typeof algorithm === "string") {
            name = algorithm.toUpperCase() === "HMAC" ? "sha256" : algorithm;
        }
        const input = binaryString(sourceBytes(data));
        const mac = __scriptgo.hmacDigest(name, binaryString(key.material), input, "hex");
        return Buffer.from(mac, "hex").buffer;
    }

    async verify(algorithm: unknown, key: CryptoKey, signature: BufferSource, data: BufferSource): Promise<boolean> {
        const expected = await this.sign(algorithm, key, data);
        return sourceBytes(signature).equals(Buffer.from(expected));
    }

    async digest(algorithm: unknown, data: BufferSource): Promise<ArrayBuffer> {
        let name = "SHA-256";
        if (typeof algorithm === "string") {
            name = algorithm;
        }
        const input = binaryString(sourceBytes(data));
        const digest = __scriptgo.hashDigest(name, input, "hex");
        return Buffer.from(digest, "hex").buffer;
    }

    async generateKey(algorithm: unknown, extractable: boolean, keyUsages: string[]): Promise<CryptoKey | CryptoKeyPair> {
        let name = "AES-GCM";
        if (typeof algorithm === "string") name = algorithm;
        const key = new CryptoKey("secret", extractable, new KeyAlgorithm(name), keyUsages);
        key.material = __scriptgo.randomBytes(32);
        return key;
    }

    async deriveKey(algorithm: unknown, baseKey: CryptoKey, derivedKeyType: unknown, extractable: boolean, keyUsages: string[]): Promise<CryptoKey> {
        const bits = await this.deriveBits(algorithm, baseKey, 256);
        let name = "AES-GCM";
        if (typeof derivedKeyType === "string") name = derivedKeyType;
        const key = new CryptoKey("secret", extractable, new KeyAlgorithm(name), keyUsages);
        key.material = sourceBytes(bits);
        return key;
    }

    async deriveBits(algorithm: unknown, baseKey: CryptoKey, length: number): Promise<ArrayBuffer> {
        if (algorithm && typeof algorithm === "object") {
            const params = algorithm as { name?: string; hash?: string; salt?: BufferSource; iterations?: number };
            if (params.name && params.name.toUpperCase() === "PBKDF2" && params.salt && params.iterations) {
                const digest: string = params.hash || "sha256";
                const salt: string = binaryString(sourceBytes(params.salt));
                const pass: string = binaryString(baseKey.material);
                const iterations: number = params.iterations;
                const keyLength: number = length / 8;
                const derived = __scriptgo.pbkdf2Sync(pass, salt, iterations, keyLength, digest);
                return derived.buffer;
            }
        }
        return new ArrayBuffer(0);
    }

    async importKey(format: string, keyData: unknown, algorithm: unknown, extractable: boolean, keyUsages: string[]): Promise<CryptoKey> {
        let name = "RAW";
        if (typeof algorithm === "string") name = algorithm;
        const key = new CryptoKey("secret", extractable, new KeyAlgorithm(name), keyUsages);
        if (format === "raw") key.material = sourceBytes(keyData as BufferSource);
        return key;
    }

    async exportKey(format: string, key: CryptoKey): Promise<unknown> {
        if (format === "raw") return key.material.buffer;
        throw new Error("Only 'raw' format is currently supported for exportKey");
    }
}

export class Crypto {
    subtle: SubtleCrypto = new SubtleCrypto();

    getRandomValues<T extends ArrayBufferView | null>(array: T): T {
        if (array === null) return array;
        const bytes = __scriptgo.randomBytes(array.byteLength);
        const view = new Uint8Array(array.buffer, array.byteOffset, array.byteLength);
        view.set(bytes);
        return array;
    }

    randomUUID(): string {
        return __scriptgo.randomUUID();
    }
}

export const webcrypto: Crypto = new Crypto();

export default {
    Crypto,
    CryptoKey,
    CryptoKeyPair,
    SubtleCrypto,
    Algorithm,
    KeyAlgorithm,
    AesDerivedKeyParams,
    AesCbcParams,
    AesCtrParams,
    AesGcmParams,
    AesKeyAlgorithm,
    AesKeyGenParams,
    EcdhKeyDeriveParams,
    EcdsaParams,
    EcKeyAlgorithm,
    EcKeyGenParams,
    EcKeyImportParams,
    Ed448Params,
    HkdfParams,
    HmacImportParams,
    HmacKeyAlgorithm,
    HmacKeyGenParams,
    Pbkdf2Params,
    RsaHashedImportParams,
    RsaHashedKeyAlgorithm,
    RsaHashedKeyGenParams,
    RsaOaepParams,
    RsaPssParams,
    webcrypto,
};
