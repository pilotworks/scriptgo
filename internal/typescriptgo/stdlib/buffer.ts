declare namespace __scriptgo {
    function bufferAlloc(size: number, fill?: string): Buffer;
    function bufferAllocUnsafe(size: number): Buffer;
    function bufferFromString(str: string, encoding?: string): Buffer;
    function bufferConcat(list: Buffer[], totalLength?: number): Buffer;
    function bufferIsBuffer(obj: unknown): boolean;
    function bufferByteLength(string: string, encoding?: string): number;
}

export interface Buffer {
    readonly length: number;
    readonly byteLength: number;
    readonly byteOffset: number;
    readonly buffer: ArrayBuffer;
    [index: number]: number;
    set(array: ArrayLike<number> | Array<number> | Uint8Array, offset?: number): void;
    subarray(begin?: number, end?: number): Buffer;
    slice(begin?: number, end?: number): Buffer;
    copy(target: Buffer | Uint8Array, targetStart?: number, sourceStart?: number, sourceEnd?: number): number;
    fill(value: number, start?: number, end?: number): this;
    equals(other: Buffer | Uint8Array): boolean;
    compare(other: Buffer | Uint8Array): number;
    indexOf(value: string, byteOffset?: number): number;
    toString(encoding?: string, start?: number, end?: number): string;

    readUInt8(offset: number): number;
    writeUInt8(value: number, offset: number): number;
    readInt8(offset: number): number;
    writeInt8(value: number, offset: number): number;
    readUInt16LE(offset: number): number;
    readUInt16BE(offset: number): number;
    writeUInt16LE(value: number, offset: number): number;
    writeUInt16BE(value: number, offset: number): number;
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
}

export class Buffer {
    static alloc(size: number, fill: string = ""): Buffer {
        return __scriptgo.bufferAlloc(size, fill);
    }
    static allocUnsafe(size: number): Buffer {
        return __scriptgo.bufferAllocUnsafe(size);
    }
    static from(str: string, encoding: string = "utf8"): Buffer {
        return __scriptgo.bufferFromString(str, encoding);
    }
    static concat(list: Buffer[], totalLength: number = -1): Buffer {
        return __scriptgo.bufferConcat(list, totalLength);
    }
    static isBuffer(obj: unknown): boolean {
        return __scriptgo.bufferIsBuffer(obj);
    }
    static byteLength(string: string, encoding: string = "utf8"): number {
        return __scriptgo.bufferByteLength(string, encoding);
    }
}
