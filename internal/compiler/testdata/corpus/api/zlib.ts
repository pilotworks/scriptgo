import {
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
    crc32
} from "node:zlib";

const zlibCallbackState = { count: 0 };

// @api: zlib.constants
// @expect: zlib_const: 0
console.log("zlib_const: " + constants.Z_NO_FLUSH);

// @api: zlib.crc32
// @expect: zlib_crc32: 3632233996
console.log("zlib_crc32: " + crc32("test"));

// @expect: zlib_crc32_bytes: 3421780262
console.log("zlib_crc32_bytes: " + crc32(new Uint8Array([49, 50, 51, 52, 53, 54, 55, 56, 57])));

// @api: zlib.deflate
deflate("hello", (err: Error | null, res: Uint8Array) => {
        if (res.length >= 0) zlibCallbackState.count++;
});

// @api: zlib.deflateSync
// @expect: zlib_deflateSync: true
const deflated = deflateSync("hello");
console.log("zlib_deflateSync: " + (deflated.length > 0));

// @api: zlib.deflateRaw
deflateRaw("hello", (err: Error | null, res: Uint8Array) => {
        if (res.length >= 0) zlibCallbackState.count++;
});

// @api: zlib.deflateRawSync
// @expect: zlib_deflateRawSync: true
console.log("zlib_deflateRawSync: " + (deflateRawSync("hello").length >= 0));

// @api: zlib.gzip
gzip("hello", (err: Error | null, res: Uint8Array) => {
        if (res.length >= 0) zlibCallbackState.count++;
});

// @api: zlib.gzipSync
// @expect: zlib_gzipSync: true
const gzipped = gzipSync("hello");
console.log("zlib_gzipSync: " + (gzipped.length > 0));

// @api: zlib.gunzip
gunzip(gzipped, (err: Error | null, res: Uint8Array) => {
        if (res.length === 5) zlibCallbackState.count++;
});

// @api: zlib.gunzipSync
// @expect: zlib_gunzipSync: true
console.log("zlib_gunzipSync: " + (gunzipSync(gzipped).length === 5));

// @api: zlib.inflate
inflate(deflated, (err: Error | null, res: Uint8Array) => {
        if (res.length === 5) zlibCallbackState.count++;
});

// @api: zlib.inflateSync
// @expect: zlib_inflateSync: true
console.log("zlib_inflateSync: " + (inflateSync(deflated).length === 5));

// @api: zlib.inflateRaw
const rawDeflated = deflateRawSync("hello");
inflateRaw(rawDeflated, (err: Error | null, res: Uint8Array) => {
        if (res.length === 5) zlibCallbackState.count++;
});

// @api: zlib.inflateRawSync
// @expect: zlib_inflateRawSync: true
console.log("zlib_inflateRawSync: " + (inflateRawSync(rawDeflated).length === 5));

// @api: zlib.unzip
unzip(gzipped, (err: Error | null, res: Uint8Array) => {
        if (res.length === 5) zlibCallbackState.count++;
});

// @api: zlib.unzipSync
// @expect: zlib_unzipSync: true
console.log("zlib_unzipSync: " + (unzipSync(gzipped).length === 5));

// @api: zlib.brotliCompress
brotliCompress("hello", (err: Error | null, res: Uint8Array) => {
        if (res.length > 0) zlibCallbackState.count++;
});

// @api: zlib.brotliCompressSync
// @expect: zlib_brotliCompressSync: true
const brotliCompressed = brotliCompressSync("hello");
console.log("zlib_brotliCompressSync: " + (brotliCompressed.length > 0));

// @api: zlib.brotliDecompress
brotliDecompress(brotliCompressed, (err: Error | null, res: Uint8Array) => {
        if (res.length === 5) zlibCallbackState.count++;
});

// @api: zlib.brotliDecompressSync
// @expect: zlib_brotliDecompressSync: true
console.log("zlib_brotliDecompressSync: " + (brotliDecompressSync(brotliCompressed).length === 5));

// @api: zlib.zstdCompress
zstdCompress("hello", (err: Error | null, res: Uint8Array) => {
        if (res.length > 0) zlibCallbackState.count++;
});

// @api: zlib.zstdCompressSync
// @expect: zlib_zstdCompressSync: true
const zstdCompressed = zstdCompressSync("hello");
console.log("zlib_zstdCompressSync: " + (zstdCompressed.length > 0));

// @api: zlib.zstdDecompress
zstdDecompress(zstdCompressed, (err: Error | null, res: Uint8Array) => {
        if (res.length === 5) zlibCallbackState.count++;
});

// @api: zlib.zstdDecompressSync
// @expect: zlib_zstdDecompressSync: true
console.log("zlib_zstdDecompressSync: " + (zstdDecompressSync(zstdCompressed).length === 5));

// @expect: zlib_callbacks_async: true
setTimeout(() => {
    console.log("zlib_callbacks_async: " + (zlibCallbackState.count === 11));
}, 0);
