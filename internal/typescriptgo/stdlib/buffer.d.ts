export interface Buffer {
    readonly length: number;
    readonly byteLength: number;
    readonly byteOffset: number;
    readonly buffer: ArrayBuffer;
    [index: number]: number;
    set(array: ArrayLike<number> | Array<number> | Uint8Array, offset?: number): void;
    toString(encoding?: string, start?: number, end?: number): string;
    subarray(begin?: number, end?: number): Buffer;
    slice(begin?: number, end?: number): Buffer;
    copy(target: Buffer | Uint8Array, targetStart?: number, sourceStart?: number, sourceEnd?: number): number;
    fill(value: number, start?: number, end?: number): this;
    equals(other: Buffer | Uint8Array): boolean;
    compare(other: Buffer | Uint8Array): number;
    indexOf(value: string, byteOffset?: number): number;

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

export interface BufferConstructor {
    alloc(size: number, fill?: number | string): Buffer;
    allocUnsafe(size: number): Buffer;
    from(str: string, encoding?: string): Buffer;
    from(array: ArrayLike<number> | Array<number> | Uint8Array | ArrayBuffer): Buffer;
    concat(list: (Buffer | Uint8Array)[], totalLength?: number): Buffer;
    isBuffer(obj: any): boolean;
    byteLength(string: string, encoding?: string): number;
    readonly prototype: Buffer;
}

export var Buffer: BufferConstructor;
