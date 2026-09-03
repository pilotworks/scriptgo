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

// @api: zlib.constants
// @expect: zlib_const: 0
console.log("zlib_const: " + constants.Z_NO_FLUSH);

// @api: zlib.crc32
// @expect: zlib_crc32: 3632233996
console.log("zlib_crc32: " + crc32("test"));

// @expect: zlib_crc32_bytes: 3421780262
console.log("zlib_crc32_bytes: " + crc32(new Uint8Array([49, 50, 51, 52, 53, 54, 55, 56, 57])));

// @api: zlib.deflate
// @expect: zlib_deflate_cb: true
deflate("hello", (err: Error | null, res: Uint8Array) => {
    console.log("zlib_deflate_cb: " + (res.length >= 0));
});

// @api: zlib.deflateSync
// @expect: zlib_deflateSync: true
const deflated = deflateSync("hello");
console.log("zlib_deflateSync: " + (deflated.length > 0));

// @api: zlib.deflateRaw
// @expect: zlib_deflateRaw_cb: true
deflateRaw("hello", (err: Error | null, res: Uint8Array) => {
    console.log("zlib_deflateRaw_cb: " + (res.length >= 0));
});

// @api: zlib.deflateRawSync
// @expect: zlib_deflateRawSync: true
console.log("zlib_deflateRawSync: " + (deflateRawSync("hello").length >= 0));

// @api: zlib.gzip
// @expect: zlib_gzip_cb: true
gzip("hello", (err: Error | null, res: Uint8Array) => {
    console.log("zlib_gzip_cb: " + (res.length >= 0));
});

// @api: zlib.gzipSync
// @expect: zlib_gzipSync: true
const gzipped = gzipSync("hello");
console.log("zlib_gzipSync: " + (gzipped.length > 0));

// @api: zlib.gunzip
// @expect: zlib_gunzip_cb: true
gunzip(gzipped, (err: Error | null, res: Uint8Array) => {
    console.log("zlib_gunzip_cb: " + (res.length === 5));
});

// @api: zlib.gunzipSync
// @expect: zlib_gunzipSync: true
console.log("zlib_gunzipSync: " + (gunzipSync(gzipped).length === 5));

// @api: zlib.inflate
// @expect: zlib_inflate_cb: true
inflate(deflated, (err: Error | null, res: Uint8Array) => {
    console.log("zlib_inflate_cb: " + (res.length === 5));
});

// @api: zlib.inflateSync
// @expect: zlib_inflateSync: true
console.log("zlib_inflateSync: " + (inflateSync(deflated).length === 5));

// @api: zlib.inflateRaw
// @expect: zlib_inflateRaw_cb: true
const rawDeflated = deflateRawSync("hello");
inflateRaw(rawDeflated, (err: Error | null, res: Uint8Array) => {
    console.log("zlib_inflateRaw_cb: " + (res.length === 5));
});

// @api: zlib.inflateRawSync
// @expect: zlib_inflateRawSync: true
console.log("zlib_inflateRawSync: " + (inflateRawSync(rawDeflated).length === 5));

// @api: zlib.unzip
// @expect: zlib_unzip_cb: true
unzip(gzipped, (err: Error | null, res: Uint8Array) => {
    console.log("zlib_unzip_cb: " + (res.length === 5));
});

// @api: zlib.unzipSync
// @expect: zlib_unzipSync: true
console.log("zlib_unzipSync: " + (unzipSync(gzipped).length === 5));

// @api: zlib.brotliCompress
// @expect: zlib_brotliCompress_cb: true
brotliCompress("hello", (err: Error | null, res: Uint8Array) => {
    console.log("zlib_brotliCompress_cb: " + (res.length > 0));
});

// @api: zlib.brotliCompressSync
// @expect: zlib_brotliCompressSync: true
const brotliCompressed = brotliCompressSync("hello");
console.log("zlib_brotliCompressSync: " + (brotliCompressed.length > 0));

// @api: zlib.brotliDecompress
// @expect: zlib_brotliDecompress_cb: true
brotliDecompress(brotliCompressed, (err: Error | null, res: Uint8Array) => {
    console.log("zlib_brotliDecompress_cb: " + (res.length === 5));
});

// @api: zlib.brotliDecompressSync
// @expect: zlib_brotliDecompressSync: true
console.log("zlib_brotliDecompressSync: " + (brotliDecompressSync(brotliCompressed).length === 5));

// @api: zlib.zstdCompress
// @expect: zlib_zstdCompress_cb: true
zstdCompress("hello", (err: Error | null, res: Uint8Array) => {
    console.log("zlib_zstdCompress_cb: " + (res.length > 0));
});

// @api: zlib.zstdCompressSync
// @expect: zlib_zstdCompressSync: true
const zstdCompressed = zstdCompressSync("hello");
console.log("zlib_zstdCompressSync: " + (zstdCompressed.length > 0));

// @api: zlib.zstdDecompress
// @expect: zlib_zstdDecompress_cb: true
zstdDecompress(zstdCompressed, (err: Error | null, res: Uint8Array) => {
    console.log("zlib_zstdDecompress_cb: " + (res.length === 5));
});

// @api: zlib.zstdDecompressSync
// @expect: zlib_zstdDecompressSync: true
console.log("zlib_zstdDecompressSync: " + (zstdDecompressSync(zstdCompressed).length === 5));
