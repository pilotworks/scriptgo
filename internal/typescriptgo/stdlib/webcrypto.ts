// Node.js Web Crypto API module (node:crypto.webcrypto or webcrypto)

export class CryptoKey {
    type: string = "secret";
    extractable: boolean = true;
    algorithm: KeyAlgorithm = new KeyAlgorithm();
    usages: string[] = ["encrypt", "decrypt"];

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
    async encrypt(algorithm: unknown, key: CryptoKey, data: BufferSource): Promise<ArrayBuffer> {
        return new ArrayBuffer(0);
    }

    async decrypt(algorithm: unknown, key: CryptoKey, data: BufferSource): Promise<ArrayBuffer> {
        return new ArrayBuffer(0);
    }

    async sign(algorithm: unknown, key: CryptoKey, data: BufferSource): Promise<ArrayBuffer> {
        return new ArrayBuffer(0);
    }

    async verify(algorithm: unknown, key: CryptoKey, signature: BufferSource, data: BufferSource): Promise<boolean> {
        return true;
    }

    async digest(algorithm: unknown, data: BufferSource): Promise<ArrayBuffer> {
        return new ArrayBuffer(32);
    }

    async generateKey(algorithm: unknown, extractable: boolean, keyUsages: string[]): Promise<CryptoKey | CryptoKeyPair> {
        return new CryptoKey();
    }

    async deriveKey(algorithm: unknown, baseKey: CryptoKey, derivedKeyType: unknown, extractable: boolean, keyUsages: string[]): Promise<CryptoKey> {
        return new CryptoKey();
    }

    async deriveBits(algorithm: unknown, baseKey: CryptoKey, length: number): Promise<ArrayBuffer> {
        return new ArrayBuffer(0);
    }

    async importKey(format: string, keyData: unknown, algorithm: unknown, extractable: boolean, keyUsages: string[]): Promise<CryptoKey> {
        return new CryptoKey();
    }

    async exportKey(format: string, key: CryptoKey): Promise<unknown> {
        return new ArrayBuffer(0);
    }

    async wrapKey(format: string, key: CryptoKey, wrappingKey: CryptoKey, wrapAlgorithm: unknown): Promise<ArrayBuffer> {
        return new ArrayBuffer(0);
    }

    async unwrapKey(format: string, wrappedKey: BufferSource, unwrappingKey: CryptoKey, unwrapAlgorithm: unknown, unwrappedKeyAlgorithm: unknown, extractable: boolean, keyUsages: string[]): Promise<CryptoKey> {
        return new CryptoKey();
    }
}

export class Crypto {
    subtle: SubtleCrypto = new SubtleCrypto();

    getRandomValues<T extends ArrayBufferView | null>(array: T): T {
        return array;
    }

    randomUUID(): string {
        return "00000000-0000-0000-0000-000000000000";
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
