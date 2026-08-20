declare namespace __scriptgo {
    function randomUUID(): string;
    function hashDigest(algorithm: string, data: string, encoding?: string): string;
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

export function createHash(algorithm: string): Hash {
    return new Hash(algorithm);
}

export function randomUUID(): string {
    return __scriptgo.randomUUID();
}

