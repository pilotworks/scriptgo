// ScriptGo Corpus: TextDecoderOptions Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: textdecoderoptions.fatal
// @api: textdecoderoptions.ignoreBOM
// @expect: true
// @expect: false
const opts: TextDecoderOptions = { fatal: true, ignoreBOM: false };
console.log(opts.fatal);
console.log(opts.ignoreBOM);
