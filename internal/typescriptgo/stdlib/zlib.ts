// Node.js Zlib module (node:zlib)

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

export class ZlibBase {
    bytesRead: number = 0;
    bytesWritten: number = 0;

    constructor() {
        this.bytesRead = 0;
        this.bytesWritten = 0;
    }

    close(callback?: unknown): void {
        if (typeof callback === "function") {
            (callback as Function)();
        }
    }

    flush(kind?: unknown, callback?: unknown): void {
        if (typeof kind === "function") {
            (kind as Function)();
        } else if (typeof callback === "function") {
            (callback as Function)();
        }
    }

    params(level: number, strategy: number, callback?: unknown): void {
        if (typeof callback === "function") {
            (callback as Function)();
        }
    }

    reset(): void {
        this.bytesRead = 0;
        this.bytesWritten = 0;
    }
}

export class Deflate extends ZlibBase {}
export class DeflateRaw extends ZlibBase {}
export class Gunzip extends ZlibBase {}
export class Gzip extends ZlibBase {}
export class Inflate extends ZlibBase {}
export class InflateRaw extends ZlibBase {}
export class Unzip extends ZlibBase {}
export class BrotliCompress extends ZlibBase {}
export class BrotliDecompress extends ZlibBase {}
export class ZstdCompress extends ZlibBase {}
export class ZstdDecompress extends ZlibBase {}

export function createDeflate(options?: unknown): Deflate { return new Deflate(); }
export function createDeflateRaw(options?: unknown): DeflateRaw { return new DeflateRaw(); }
export function createGunzip(options?: unknown): Gunzip { return new Gunzip(); }
export function createGzip(options?: unknown): Gzip { return new Gzip(); }
export function createInflate(options?: unknown): Inflate { return new Inflate(); }
export function createInflateRaw(options?: unknown): InflateRaw { return new InflateRaw(); }
export function createUnzip(options?: unknown): Unzip { return new Unzip(); }
export function createBrotliCompress(options?: unknown): BrotliCompress { return new BrotliCompress(); }
export function createBrotliDecompress(options?: unknown): BrotliDecompress { return new BrotliDecompress(); }
export function createZstdCompress(options?: unknown): ZstdCompress { return new ZstdCompress(); }
export function createZstdDecompress(options?: unknown): ZstdDecompress { return new ZstdDecompress(); }

function _mockBuffer(data: unknown): Uint8Array {
    if (typeof data === "string") {
        const str = data as string;
        const res = new Uint8Array(str.length);
        for (let i = 0; i < str.length; i++) {
            res[i] = str.charCodeAt(i);
        }
        return res;
    }
    return new Uint8Array(0);
}

export function deflate(buf: unknown, options?: unknown, callback?: unknown): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        (cb as Function)(null, _mockBuffer(buf));
    }
}

export function deflateSync(buf: unknown, options?: unknown): Uint8Array {
    return _mockBuffer(buf);
}

export function deflateRaw(buf: unknown, options?: unknown, callback?: unknown): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        (cb as Function)(null, _mockBuffer(buf));
    }
}

export function deflateRawSync(buf: unknown, options?: unknown): Uint8Array {
    return _mockBuffer(buf);
}

export function gunzip(buf: unknown, options?: unknown, callback?: unknown): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        (cb as Function)(null, _mockBuffer(buf));
    }
}

export function gunzipSync(buf: unknown, options?: unknown): Uint8Array {
    return _mockBuffer(buf);
}

export function gzip(buf: unknown, options?: unknown, callback?: unknown): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        (cb as Function)(null, _mockBuffer(buf));
    }
}

export function gzipSync(buf: unknown, options?: unknown): Uint8Array {
    return _mockBuffer(buf);
}

export function inflate(buf: unknown, options?: unknown, callback?: unknown): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        (cb as Function)(null, _mockBuffer(buf));
    }
}

export function inflateSync(buf: unknown, options?: unknown): Uint8Array {
    return _mockBuffer(buf);
}

export function inflateRaw(buf: unknown, options?: unknown, callback?: unknown): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        (cb as Function)(null, _mockBuffer(buf));
    }
}

export function inflateRawSync(buf: unknown, options?: unknown): Uint8Array {
    return _mockBuffer(buf);
}

export function unzip(buf: unknown, options?: unknown, callback?: unknown): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        (cb as Function)(null, _mockBuffer(buf));
    }
}

export function unzipSync(buf: unknown, options?: unknown): Uint8Array {
    return _mockBuffer(buf);
}

export function brotliCompress(buf: unknown, options?: unknown, callback?: unknown): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        (cb as Function)(null, _mockBuffer(buf));
    }
}

export function brotliCompressSync(buf: unknown, options?: unknown): Uint8Array {
    return _mockBuffer(buf);
}

export function brotliDecompress(buf: unknown, options?: unknown, callback?: unknown): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        (cb as Function)(null, _mockBuffer(buf));
    }
}

export function brotliDecompressSync(buf: unknown, options?: unknown): Uint8Array {
    return _mockBuffer(buf);
}

export function zstdCompress(buf: unknown, options?: unknown, callback?: unknown): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        (cb as Function)(null, _mockBuffer(buf));
    }
}

export function zstdCompressSync(buf: unknown, options?: unknown): Uint8Array {
    return _mockBuffer(buf);
}

export function zstdDecompress(buf: unknown, options?: unknown, callback?: unknown): void {
    const cb = typeof options === "function" ? options : callback;
    if (typeof cb === "function") {
        (cb as Function)(null, _mockBuffer(buf));
    }
}

export function zstdDecompressSync(buf: unknown, options?: unknown): Uint8Array {
    return _mockBuffer(buf);
}

export function crc32(data: unknown, value?: number): number {
    return 0;
}

export default {
    constants,
    ZlibBase,
    Deflate,
    DeflateRaw,
    Gunzip,
    Gzip,
    Inflate,
    InflateRaw,
    Unzip,
    BrotliCompress,
    BrotliDecompress,
    ZstdCompress,
    ZstdDecompress,
    createDeflate,
    createDeflateRaw,
    createGunzip,
    createGzip,
    createInflate,
    createInflateRaw,
    createUnzip,
    createBrotliCompress,
    createBrotliDecompress,
    createZstdCompress,
    createZstdDecompress,
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
