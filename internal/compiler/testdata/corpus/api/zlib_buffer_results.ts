// @expect: xin chào 9 true
// @expect: hello hello hello
// @expect: brötli
// @expect: <Buffer 61 62 63>
// @expect: null cb
// zlib sync functions and callbacks return a Buffer as in Node.js, so
// toString() decodes text and console.log shows <Buffer ...>.
import { gzipSync, gunzipSync, deflateSync, inflateSync, gzip, brotliCompressSync, brotliDecompressSync } from "node:zlib";
const out = gunzipSync(gzipSync(Buffer.from("xin chào")));
console.log(out.toString(), out.length, Buffer.isBuffer(out));
console.log(inflateSync(deflateSync("hello hello hello")).toString("utf8"));
console.log(brotliDecompressSync(brotliCompressSync("brötli")).toString());
console.log(gunzipSync(gzipSync("abc")));
gzip("cb", (err, res) => console.log(err, gunzipSync(res).toString()));
