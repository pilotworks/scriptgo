// Node.js Zlib module (node:zlib)

declare namespace __scriptgo {
    function zlibTransformString(data: string, mode: number): Uint8Array;
    function zlibTransformBuffer(data: Uint8Array, mode: number): Uint8Array;
}

export const constants = {
    Z_NO_FLUSH: 0,
    Z_PARTIAL_FLUSH: 1,
    Z_SYNC_FLUSH: 2,
    Z_FULL_FLUSH: 3,
    Z_FINISH: 4,
    Z_BLOCK: 5,
    Z_TREES: 6,
    Z_OK: 0,
    Z_STREAM_END: 1,
    Z_NEED_DICT: 2,
    Z_ERRNO: -1,
    Z_STREAM_ERROR: -2,
    Z_DATA_ERROR: -3,
    Z_MEM_ERROR: -4,
    Z_BUF_ERROR: -5,
    Z_VERSION_ERROR: -6,
    Z_NO_COMPRESSION: 0,
    Z_BEST_SPEED: 1,
    Z_BEST_COMPRESSION: 9,
    Z_DEFAULT_COMPRESSION: -1,
    Z_FILTERED: 1,
    Z_HUFFMAN_ONLY: 2,
    Z_RLE: 3,
    Z_FIXED: 4,
    Z_DEFAULT_STRATEGY: 0,
    BROTLI_OPERATION_PROCESS: 0,
    BROTLI_OPERATION_FLUSH: 1,
    BROTLI_OPERATION_FINISH: 2,
    BROTLI_OPERATION_EMIT_METADATA: 3,
    BROTLI_PARAM_MODE: 0,
    BROTLI_PARAM_QUALITY: 1,
    BROTLI_PARAM_LGWIN: 2,
    BROTLI_PARAM_LGBLOCK: 3,
    BROTLI_PARAM_DISABLE_LITERAL_CONTEXT_MODELING: 4,
    BROTLI_PARAM_SIZE_HINT: 5,
    BROTLI_PARAM_LARGE_WINDOW: 6,
    BROTLI_PARAM_NPOSTFIX: 7,
    BROTLI_PARAM_NDIRECT: 8,
    BROTLI_DECODER_RESULT_ERROR: 0,
    BROTLI_DECODER_RESULT_SUCCESS: 1,
    BROTLI_DECODER_RESULT_NEEDS_MORE_INPUT: 2,
    BROTLI_DECODER_RESULT_NEEDS_MORE_OUTPUT: 3,
};


export interface ZlibOptions {
    flush?: number;
    finishFlush?: number;
    chunkSize?: number;
    windowBits?: number;
    level?: number;
    memLevel?: number;
    strategy?: number;
    dictionary?: string | Uint8Array;
    info?: boolean;
    maxOutputLength?: number;
}

export interface BrotliOptions extends ZlibOptions {
    params?: Record<number, number>;
}

export interface ZstdOptions extends ZlibOptions {
    params?: Record<number, number>;
}

export type ZlibCallback = (error: Error | null, result: Uint8Array) => void;

type ZlibInput = string | Uint8Array;

function _transform(data: ZlibInput, mode: number): Uint8Array {
    if (typeof data === "string") {
        return __scriptgo.zlibTransformString(data, mode);
    }
    return __scriptgo.zlibTransformBuffer(data, mode);
}

function _scheduleTransform(callback: ZlibCallback, data: ZlibInput, mode: number): void {
    Promise.resolve().then(() => callback(null, _transform(data, mode)));
}

function _crc32Bytes(data: Uint8Array, initial: number): number {
    let crc = (initial ^ -1) >>> 0;
    for (let i = 0; i < data.length; i++) {
        crc = crc ^ data[i];
        for (let bit = 0; bit < 8; bit++) {
            const mask = -(crc & 1);
            crc = (crc >>> 1) ^ (0xEDB88320 & mask);
        }
    }
    return (crc ^ -1) >>> 0;
}

function _crc32Input(data: ZlibInput): Uint8Array {
    if (typeof data === "string") {
        const result = new Uint8Array(data.length);
        for (let i = 0; i < data.length; i++) {
            result[i] = data.charCodeAt(i) & 0xFF;
        }
        return result;
    }
    return data;
}

export function deflate(buf: ZlibInput, options?: ZlibOptions | ZlibCallback, callback?: ZlibCallback): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        _scheduleTransform(cb, buf, 0);
    }
}

export function deflateSync(buf: ZlibInput, options?: ZlibOptions): Uint8Array {
    return _transform(buf, 0);
}

export function deflateRaw(buf: ZlibInput, options?: ZlibOptions | ZlibCallback, callback?: ZlibCallback): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        _scheduleTransform(cb, buf, 1);
    }
}

export function deflateRawSync(buf: ZlibInput, options?: ZlibOptions): Uint8Array {
    return _transform(buf, 1);
}

export function gunzip(buf: ZlibInput, options?: ZlibOptions | ZlibCallback, callback?: ZlibCallback): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        _scheduleTransform(cb, buf, 5);
    }
}

export function gunzipSync(buf: ZlibInput, options?: ZlibOptions): Uint8Array {
    return _transform(buf, 5);
}

export function gzip(buf: ZlibInput, options?: ZlibOptions | ZlibCallback, callback?: ZlibCallback): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        _scheduleTransform(cb, buf, 2);
    }
}

export function gzipSync(buf: ZlibInput, options?: ZlibOptions): Uint8Array {
    return _transform(buf, 2);
}

export function inflate(buf: ZlibInput, options?: ZlibOptions | ZlibCallback, callback?: ZlibCallback): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        _scheduleTransform(cb, buf, 3);
    }
}

export function inflateSync(buf: ZlibInput, options?: ZlibOptions): Uint8Array {
    return _transform(buf, 3);
}

export function inflateRaw(buf: ZlibInput, options?: ZlibOptions | ZlibCallback, callback?: ZlibCallback): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        _scheduleTransform(cb, buf, 4);
    }
}

export function inflateRawSync(buf: ZlibInput, options?: ZlibOptions): Uint8Array {
    return _transform(buf, 4);
}

export function unzip(buf: ZlibInput, options?: ZlibOptions | ZlibCallback, callback?: ZlibCallback): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        _scheduleTransform(cb, buf, 5);
    }
}

export function unzipSync(buf: ZlibInput, options?: ZlibOptions): Uint8Array {
    return _transform(buf, 5);
}

export function brotliCompress(buf: ZlibInput, options?: ZlibOptions | ZlibCallback, callback?: ZlibCallback): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        _scheduleTransform(cb, buf, 6);
    }
}

export function brotliCompressSync(buf: ZlibInput, options?: ZlibOptions): Uint8Array {
    return _transform(buf, 6);
}

export function brotliDecompress(buf: ZlibInput, options?: ZlibOptions | ZlibCallback, callback?: ZlibCallback): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        _scheduleTransform(cb, buf, 7);
    }
}

export function brotliDecompressSync(buf: ZlibInput, options?: ZlibOptions): Uint8Array {
    return _transform(buf, 7);
}

export function zstdCompress(buf: ZlibInput, options?: ZlibOptions | ZlibCallback, callback?: ZlibCallback): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        _scheduleTransform(cb, buf, 8);
    }
}

export function zstdCompressSync(buf: ZlibInput, options?: ZlibOptions): Uint8Array {
    return _transform(buf, 8);
}

export function zstdDecompress(buf: ZlibInput, options?: ZlibOptions | ZlibCallback, callback?: ZlibCallback): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        _scheduleTransform(cb, buf, 9);
    }
}

export function zstdDecompressSync(buf: ZlibInput, options?: ZlibOptions): Uint8Array {
    return _transform(buf, 9);
}

export function crc32(data: ZlibInput, value?: number): number {
    const initial = value === undefined ? 0 : value;
    return _crc32Bytes(_crc32Input(data), initial);
}

export default {
    constants,
    deflate,
    deflateSync,
    deflateRaw,
    deflateRawSync,
    gunzip,
    gunzipSync,
    gzip,
    gzipSync,
    inflate,
    inflateSync,
    inflateRaw,
    inflateRawSync,
    unzip,
    unzipSync,
    brotliCompress,
    brotliCompressSync,
    brotliDecompress,
    brotliDecompressSync,
    zstdCompress,
    zstdCompressSync,
    zstdDecompress,
    zstdDecompressSync,
    crc32,
};
