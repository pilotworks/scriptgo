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

function bufferToArrayBuffer(buf: Buffer): ArrayBuffer {
    const ab = new ArrayBuffer(buf.length);
    const u8 = new Uint8Array(ab);
    for (let i = 0; i < buf.length; i++) {
        u8[i] = buf[i];
    }
    return ab;
}

export class SubtleCrypto {
    async sign(algorithm: unknown, key: CryptoKey, data: BufferSource): Promise<ArrayBuffer> {
        let name = key.algorithm.name;
        if (typeof algorithm === "string") {
            name = algorithm.toUpperCase() === "HMAC" ? "sha256" : algorithm;
        }
        const input = binaryString(sourceBytes(data));
        const mac = __scriptgo.hmacDigest(name, binaryString(key.material), input, "hex");
        return bufferToArrayBuffer(Buffer.from(mac, "hex"));
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
        return bufferToArrayBuffer(Buffer.from(digest, "hex"));
    }

    async deriveBits(algorithm: unknown, baseKey: CryptoKey, length: number): Promise<ArrayBuffer> {
        let name = "";
        let salt = "salt";
        let iterations = 1000;
        let digest = "sha256";
        if (typeof algorithm === "string") {
            name = algorithm;
        } else if (algorithm && typeof algorithm === "object") {
            const params = algorithm as { name?: string; hash?: string; salt?: BufferSource; iterations?: number };
            if (params.name) name = params.name;
            if (params.hash) digest = params.hash;
            if (params.iterations) iterations = params.iterations;
            if (params.salt) salt = binaryString(sourceBytes(params.salt));
        }
        if (name.toUpperCase() === "PBKDF2" || name.toUpperCase() === "HKDF") {
            const pass = binaryString(baseKey.material);
            const keyLength = length > 0 ? length / 8 : 16;
            const derived = __scriptgo.pbkdf2Sync(pass, salt, iterations, keyLength, digest);
            return bufferToArrayBuffer(derived);
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
        if (format === "raw") return bufferToArrayBuffer(key.material);
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
    webcrypto,
};
