import {
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

// @api: zlib.zlib.ZlibBase
// @expect: zlib_base_inst: true
const base = new ZlibBase();
console.log("zlib_base_inst: " + (base instanceof ZlibBase));

// @api: zlib.ZlibBase.bytesRead
// @expect: zlib_bytesRead: 0
console.log("zlib_bytesRead: " + base.bytesRead);

// @api: zlib.ZlibBase.bytesWritten
// @expect: zlib_bytesWritten: 0
console.log("zlib_bytesWritten: " + base.bytesWritten);

// @api: zlib.ZlibBase.close
// @expect: zlib_close: true
base.close();
console.log("zlib_close: true");

// @api: zlib.ZlibBase.flush
// @expect: zlib_flush: true
base.flush();
console.log("zlib_flush: true");

// @api: zlib.ZlibBase.params
// @expect: zlib_params: true
base.params(1, 0);
console.log("zlib_params: true");

// @api: zlib.ZlibBase.reset
// @expect: zlib_reset: true
base.reset();
console.log("zlib_reset: true");

// @api: zlib.zlib.Deflate
// @api: zlib.createDeflate
// @expect: zlib_deflate_inst: true
const d = createDeflate();
console.log("zlib_deflate_inst: " + (d instanceof Deflate));

// @api: zlib.zlib.DeflateRaw
// @api: zlib.createDeflateRaw
// @expect: zlib_deflateRaw_inst: true
const dr = createDeflateRaw();
console.log("zlib_deflateRaw_inst: " + (dr instanceof DeflateRaw));

// @api: zlib.zlib.Gunzip
// @api: zlib.createGunzip
// @expect: zlib_gunzip_inst: true
const gz = createGunzip();
console.log("zlib_gunzip_inst: " + (gz instanceof Gunzip));

// @api: zlib.zlib.Gzip
// @api: zlib.createGzip
// @expect: zlib_gzip_inst: true
const g = createGzip();
console.log("zlib_gzip_inst: " + (g instanceof Gzip));

// @api: zlib.zlib.Inflate
// @api: zlib.createInflate
// @expect: zlib_inflate_inst: true
const inf = createInflate();
console.log("zlib_inflate_inst: " + (inf instanceof Inflate));

// @api: zlib.zlib.InflateRaw
// @api: zlib.createInflateRaw
// @expect: zlib_inflateRaw_inst: true
const infr = createInflateRaw();
console.log("zlib_inflateRaw_inst: " + (infr instanceof InflateRaw));

// @api: zlib.zlib.Unzip
// @api: zlib.createUnzip
// @expect: zlib_unzip_inst: true
const u = createUnzip();
console.log("zlib_unzip_inst: " + (u instanceof Unzip));

// @api: zlib.zlib.BrotliCompress
// @api: zlib.createBrotliCompress
// @expect: zlib_brotliCompress_inst: true
const bc = createBrotliCompress();
console.log("zlib_brotliCompress_inst: " + (bc instanceof BrotliCompress));

// @api: zlib.zlib.BrotliDecompress
// @api: zlib.createBrotliDecompress
// @expect: zlib_brotliDecompress_inst: true
const bd = createBrotliDecompress();
console.log("zlib_brotliDecompress_inst: " + (bd instanceof BrotliDecompress));

// @api: zlib.zlib.ZstdCompress
// @api: zlib.createZstdCompress
// @expect: zlib_zstdCompress_inst: true
const zc = createZstdCompress();
console.log("zlib_zstdCompress_inst: " + (zc instanceof ZstdCompress));

// @api: zlib.zlib.ZstdDecompress
// @api: zlib.createZstdDecompress
// @expect: zlib_zstdDecompress_inst: true
const zd = createZstdDecompress();
console.log("zlib_zstdDecompress_inst: " + (zd instanceof ZstdDecompress));

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
