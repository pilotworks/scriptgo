// ScriptGo Corpus: TextEncoder Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: textencoder.encode
// @api: textencoder.encoding
// @api: textencoder.constructor
// @expect: utf-8
// @expect: 5
const encoder = new TextEncoder();
console.log(encoder.encoding);
const u8 = encoder.encode("hello");
console.log(u8.length);

// @api: textencoder.encodeInto
// @expect: 5
// @expect: 5
const dest = new Uint8Array(10);
const res = encoder.encodeInto("hello", dest);
console.log(res.read);
console.log(res.written);
