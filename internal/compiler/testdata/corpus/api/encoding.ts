// ScriptGo Corpus: Encoding Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: encoding.atob
// @expect: hello
console.log(atob("aGVsbG8="));

// @api: encoding.btoa
// @expect: aGVsbG8=
console.log(btoa("hello"));

// @api: encoding.decode
// @expect: hello decoder
const encoder_encoding_decode_2 = new TextEncoder();
const decoder_encoding_decode_2 = new TextDecoder();
const u8_encoding_decode_2 = encoder_encoding_decode_2.encode("hello decoder");
console.log(decoder_encoding_decode_2.decode(u8_encoding_decode_2));

// @api: encoding.encode
// @expect: 2
// @expect: 104
const encoder_encoding_encode_3 = new TextEncoder();
const u8_encoding_encode_3: Uint8Array = encoder_encoding_encode_3.encode("hi");
console.log(u8_encoding_encode_3.length);
console.log(u8_encoding_encode_3[0]);
