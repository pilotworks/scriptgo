declare namespace __scriptgo {
    function bufferAlloc(size: number, fill?: string): Buffer;
    function bufferAllocUnsafe(size: number): Buffer;
    function bufferFromString(str: string, encoding?: string): Buffer;
    function bufferFromArray(array: ArrayLike<number> | Array<number> | Uint8Array): Buffer;
    function bufferFromArrayBuffer(array: ArrayBuffer): Buffer;
    function bufferConcat(list: (Buffer | Uint8Array)[], totalLength?: number): Buffer;
    function bufferIsBuffer(obj: unknown): boolean;
    function bufferByteLength(string: string, encoding?: string): number;
    function atob(data: string): string;
    function btoa(data: string): string;
}

export interface Buffer extends Uint8Array {
    copy(target: Buffer | Uint8Array, targetStart?: number, sourceStart?: number, sourceEnd?: number): number;
    equals(other: Buffer | Uint8Array): boolean;
    compare(other: Buffer | Uint8Array): number;
    toString(encoding?: string, start?: number, end?: number): string;

    readUInt8(offset: number): number;
    writeUInt8(value: number, offset: number): number;
    readInt8(offset: number): number;
    writeInt8(value: number, offset: number): number;
    readUInt16LE(offset: number): number;
    readUInt16BE(offset: number): number;
    writeUInt16LE(value: number, offset: number): number;
    writeUInt16BE(value: number, offset: number): number;
    readInt16LE(offset: number): number;
    readInt16BE(offset: number): number;
    writeInt16LE(value: number, offset: number): number;
    writeInt16BE(value: number, offset: number): number;
    readUInt32LE(offset: number): number;
    readUInt32BE(offset: number): number;
    writeUInt32LE(value: number, offset: number): number;
    writeUInt32BE(value: number, offset: number): number;
    readInt32LE(offset: number): number;
    readInt32BE(offset: number): number;
    writeInt32LE(value: number, offset: number): number;
    writeInt32BE(value: number, offset: number): number;
    readFloatLE(offset: number): number;
    readFloatBE(offset: number): number;
    writeFloatLE(value: number, offset: number): number;
    writeFloatBE(value: number, offset: number): number;
    readDoubleLE(offset: number): number;
    readDoubleBE(offset: number): number;
    writeDoubleLE(value: number, offset: number): number;
    writeDoubleBE(value: number, offset: number): number;

    readBigInt64BE(offset?: number): bigint;
    readBigInt64LE(offset?: number): bigint;
    readBigUInt64BE(offset?: number): bigint;
    readBigUInt64LE(offset?: number): bigint;
    writeBigInt64BE(value: bigint, offset?: number): number;
    writeBigInt64LE(value: bigint, offset?: number): number;
    writeBigUInt64BE(value: bigint, offset?: number): number;
    writeBigUInt64LE(value: bigint, offset?: number): number;

    readIntBE(offset: number, byteLength: number): number;
    readIntLE(offset: number, byteLength: number): number;
    readUIntBE(offset: number, byteLength: number): number;
    readUIntLE(offset: number, byteLength: number): number;
    writeIntBE(value: number, offset: number, byteLength: number): number;
    writeIntLE(value: number, offset: number, byteLength: number): number;
    writeUIntBE(value: number, offset: number, byteLength: number): number;
    writeUIntLE(value: number, offset: number, byteLength: number): number;

    toJSON(): { type: "Buffer"; data: number[] };
    swap16(): this;
    swap32(): this;
    swap64(): this;
    write(string: string, offset?: number, length?: number, encoding?: string): number;

    readonly parent: ArrayBuffer;
}

export class Buffer {
    static alloc(size: number, fill: string = ""): Buffer {
        return __scriptgo.bufferAlloc(size, fill);
    }
    static allocUnsafe(size: number): Buffer {
        return __scriptgo.bufferAllocUnsafe(size);
    }
    static from(str: string, encoding?: string): Buffer;
    static from(array: ArrayLike<number> | Array<number> | Uint8Array | ArrayBuffer): Buffer;
    static from(value: string | ArrayLike<number> | Array<number> | Uint8Array | ArrayBuffer, encoding: string = "utf8"): Buffer {
        if (typeof value === "string") {
            return __scriptgo.bufferFromString(value, encoding);
        }
        if (value instanceof ArrayBuffer) {
            return __scriptgo.bufferFromArrayBuffer(value);
        }
        return __scriptgo.bufferFromArray(value);
    }
    static concat(list: (Buffer | Uint8Array)[], totalLength: number = -1): Buffer {
        return __scriptgo.bufferConcat(list, totalLength);
    }
    static isBuffer(obj: unknown): obj is Buffer {
        return __scriptgo.bufferIsBuffer(obj);
    }
    static byteLength(string: string, encoding: string = "utf8"): number {
        return __scriptgo.bufferByteLength(string, encoding);
    }
    static compare(buf1: Buffer, buf2: Buffer): number {
        return buf1.compare(buf2);
    }
    static isEncoding(encoding: string): boolean {
        return true;
    }
    static copyBytesFrom(view: Uint8Array, offset?: number, length?: number): Buffer {
        return Buffer.from(view);
    }
    static poolSize: number = 8192;
}

type BlobPart = string | Buffer | Uint8Array | ArrayBuffer | Blob;

export class Blob {
    size: number = 0;
    type: string = "";
    private _bytes: Buffer = Buffer.alloc(0);

    constructor(sources: BlobPart[] = [], options?: { type?: string }) {
        const parts: Buffer[] = [];
        for (let i = 0; i < sources.length; i++) {
            const source = sources[i];
            if (source instanceof Blob) {
                parts.push(source._bytes);
            } else if (typeof source === "string") {
                parts.push(Buffer.from(source));
            } else {
                parts.push(Buffer.from(source));
            }
        }
        this._bytes = Buffer.concat(parts);
        this.size = this._bytes.length;
        if (options && options.type) {
            this.type = options.type.toLowerCase();
        }
    }

    async arrayBuffer(): Promise<ArrayBuffer> {
        const result = new ArrayBuffer(this._bytes.length);
        new Uint8Array(result).set(this._bytes);
        return result;
    }

    slice(start: number = 0, end: number = this.size, contentType: string = ""): Blob {
        return new Blob([this._bytes.slice(start, end)], { type: contentType });
    }

    async text(): Promise<string> {
        return this._bytes.toString();
    }

    stream(): unknown {
        return null;
    }
}

export class File extends Blob {
    name: string = "";
    lastModified: number = 0;
    constructor(fileBits: BlobPart[], fileName: string, options?: { type?: string; lastModified?: number }) {
        super(fileBits, options);
        this.name = fileName;
        if (options) {
            if (options.lastModified) {
                this.lastModified = options.lastModified;
            }
        }
    }
}

export const constants = {
    MAX_LENGTH: 2147483647,
    MAX_STRING_LENGTH: 536870888,
};

export const MAX_LENGTH: number = constants.MAX_LENGTH;
export const MAX_STRING_LENGTH: number = constants.MAX_STRING_LENGTH;
export const kMaxLength: number = constants.MAX_LENGTH;
export const kStringMaxLength: number = constants.MAX_STRING_LENGTH;
export const INSPECT_MAX_BYTES: number = 50;
export class SlowBuffer {
    constructor(size: number) {}
}

export function atob(data: string): string {
    return __scriptgo.atob(data);
}

export function btoa(data: string): string {
    return __scriptgo.btoa(data);
}

export function isAscii(input: Buffer | Uint8Array): boolean {
    return true;
}

export function isUtf8(input: Buffer | Uint8Array): boolean {
    return true;
}

export function resolveObjectURL(id: string): unknown {
    return null;
}

export function transcode(source: Uint8Array, fromEnc: string, toEnc: string): Buffer {
    return Buffer.from(source);
}
