import { Buffer } from "node:buffer";
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
    crc32,
    createDeflate,
    createDeflateRaw,
    createGzip,
    createGunzip,
    createInflate,
    createInflateRaw,
    createUnzip,
    createBrotliCompress,
    createBrotliDecompress,
    createZstdCompress,
    createZstdDecompress
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

// @api: zlib.ZlibBase
const testBase = createGzip();

// @api: zlib.ZlibBase.bytesRead
// @api: zlib.ZlibBase.bytesWritten
// @expect: zlib_base_bytes: true
console.log("zlib_base_bytes: " + (testBase.bytesWritten === 0 && (testBase.bytesRead === 0 || testBase.bytesRead === undefined)));

// @api: zlib.ZlibBase.reset
testBase.reset();

// @api: zlib.ZlibBase.params
const baseParams = createGzip();
baseParams.params(1, 0, () => {
    zlibCallbackState.count++;
});

// @api: zlib.ZlibBase.flush
const baseFlush = createGzip();
baseFlush.flush(0, () => {
    zlibCallbackState.count++;
});

// @api: zlib.ZlibBase.close
const baseClose = createGzip();
baseClose.close(() => {
    zlibCallbackState.count++;
});

// @api: zlib.createDeflate
const streamDeflate = createDeflate();

// @api: zlib.createDeflateRaw
const streamDeflateRaw = createDeflateRaw();

// @api: zlib.createGzip
const streamGzip = createGzip();

// @api: zlib.createGunzip
const streamGunzip = createGunzip();

// @api: zlib.createInflate
const streamInflate = createInflate();

// @api: zlib.createInflateRaw
const streamInflateRaw = createInflateRaw();

// @api: zlib.createUnzip
const streamUnzip = createUnzip();

// @api: zlib.createBrotliCompress
const streamBrotliCompress = createBrotliCompress();

// @api: zlib.createBrotliDecompress
const streamBrotliDecompress = createBrotliDecompress();

// @api: zlib.createZstdCompress
const streamZstdCompress = createZstdCompress();

// @api: zlib.createZstdDecompress
const streamZstdDecompress = createZstdDecompress();

// @expect: zlib_factory_streams: true
console.log("zlib_factory_streams: " + (
    typeof streamDeflate === "object" &&
    typeof streamDeflateRaw === "object" &&
    typeof streamGzip === "object" &&
    typeof streamGunzip === "object" &&
    typeof streamInflate === "object" &&
    typeof streamInflateRaw === "object" &&
    typeof streamUnzip === "object" &&
    typeof streamBrotliCompress === "object" &&
    typeof streamBrotliDecompress === "object" &&
    typeof streamZstdCompress === "object" &&
    typeof streamZstdDecompress === "object"
));

// @expect: zlib_stream_pipeline: hello stream
streamGzip.pipe(streamGunzip);
let streamedData = "";
streamGunzip.on("data", (chunk: Buffer) => {
    streamedData += chunk.toString();
});
streamGunzip.on("end", () => {
    console.log("zlib_stream_pipeline: " + streamedData);
    zlibCallbackState.count++;
});
streamGzip.write("hello stream");
streamGzip.end();

// @expect: zlib_callbacks_async: true
const zlibCallbackPoller = setInterval(() => {
    if (zlibCallbackState.count === 15) {
        clearInterval(zlibCallbackPoller);
        console.log("zlib_callbacks_async: true");
    }
}, 10);
