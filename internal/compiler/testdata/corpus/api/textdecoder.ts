// ScriptGo Corpus: TextDecoder Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: textdecoder.decode
// @api: textdecoder.encoding
// @api: textdecoder.fatal
// @api: textdecoder.ignoreBOM
// @api: textdecoder.constructor
// @expect: utf-8
// @expect: false
// @expect: false
// @expect: hello world
const decoder = new TextDecoder("utf-8");
console.log(decoder.encoding);
console.log(decoder.fatal);
console.log(decoder.ignoreBOM);

const encoder = new TextEncoder();
const bytes = encoder.encode("hello world");
console.log(decoder.decode(bytes));
