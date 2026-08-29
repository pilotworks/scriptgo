declare namespace __scriptgo {
    function randomUUID(): string;
    function hashDigest(algorithm: string, data: string, encoding?: string): string;
    function hmacDigest(algorithm: string, key: string, data: string, encoding?: string): string;
    function randomBytes(size: number): Buffer;
    function randomInt(min: number, max: number): number;
    function randomFill(buffer: Buffer, offset?: number, size?: number): Buffer;
    function timingSafeEqual(a: Buffer, b: Buffer): boolean;
    function pbkdf2Sync(password: string, salt: string, iterations: number, keylen: number, digest?: string): Buffer;
}

export class Hash {
    private algorithm: string;
    private data: string;

    constructor(algorithm: string) {
        this.algorithm = algorithm;
        this.data = "";
    }

    update(data: string): Hash {
        this.data = this.data + data;
        return this;
    }

    digest(encoding: string = "hex"): string {
        return __scriptgo.hashDigest(this.algorithm, this.data, encoding);
    }
}

export class Hmac {
    private algorithm: string;
    private key: string;
    private data: string;

    constructor(algorithm: string, key: string) {
        this.algorithm = algorithm;
        this.key = key;
        this.data = "";
    }

    update(data: string): Hmac {
        this.data = this.data + data;
        return this;
    }

    digest(encoding: string = "hex"): string {
        return __scriptgo.hmacDigest(this.algorithm, this.key, this.data, encoding);
    }
}

export function createHash(algorithm: string): Hash {
    return new Hash(algorithm);
}

export function createHmac(algorithm: string, key: string): Hmac {
    return new Hmac(algorithm, key);
}

export function randomUUID(): string {
    return __scriptgo.randomUUID();
}

export function randomBytes(size: number): Buffer {
    return __scriptgo.randomBytes(size);
}

export function randomInt(min: number, max?: number): number {
    if (max === undefined) {
        return __scriptgo.randomInt(0, min);
    }
    return __scriptgo.randomInt(min, max);
}

export function randomFillSync(buffer: Buffer, offset: number = 0, size?: number): Buffer {
    return __scriptgo.randomFill(buffer, offset, size);
}

export function timingSafeEqual(a: Buffer, b: Buffer): boolean {
    return __scriptgo.timingSafeEqual(a, b);
}

export function pbkdf2Sync(password: string, salt: string, iterations: number, keylen: number, digest: string = "sha1"): Buffer {
    return __scriptgo.pbkdf2Sync(password, salt, iterations, keylen, digest);
}

export function getHashes(): string[] {
    return ["md5", "sha1", "sha256", "sha512"];
}

export const constants = {
    RSA_PKCS1_PADDING: 1,
    RSA_SSLV23_PADDING: 2,
    RSA_NO_PADDING: 3,
    RSA_PKCS1_OAEP_PADDING: 4,
    RSA_X931_PADDING: 5,
    RSA_PKCS1_PSS_PADDING: 6,
};

export default {
    Hash,
    Hmac,
    createHash,
    createHmac,
    randomUUID,
    randomBytes,
    randomInt,
    randomFillSync,
    timingSafeEqual,
    pbkdf2Sync,
    getHashes,
    constants,
};


