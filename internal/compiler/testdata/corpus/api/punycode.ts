import punycode, { encode, decode, toASCII, toUnicode, ucs2, version } from "node:punycode";
import punycodeBare from "punycode";

// @api: punycode.toASCII
// @expect: toASCII mañana.com: xn--maana-pta.com
// @expect: toASCII bücher.de: xn--bcher-kva.de
// @expect: toASCII ascii.com: ascii.com
console.log("toASCII mañana.com:", toASCII("mañana.com"));
console.log("toASCII bücher.de:", toASCII("bücher.de"));
console.log("toASCII ascii.com:", toASCII("ascii.com"));

// @api: punycode.toUnicode
// @expect: toUnicode xn--maana-pta.com: mañana.com
// @expect: toUnicode xn--bcher-kva.de: bücher.de
// @expect: toUnicode ascii.com: ascii.com
console.log("toUnicode xn--maana-pta.com:", toUnicode("xn--maana-pta.com"));
console.log("toUnicode xn--bcher-kva.de:", toUnicode("xn--bcher-kva.de"));
console.log("toUnicode ascii.com:", toUnicode("ascii.com"));

// @api: punycode.encode
// @api: punycode.decode
// @expect: encode mañana: maana-pta
// @expect: decode maana-pta: mañana
// @expect: encode bücher: bcher-kva
// @expect: decode bcher-kva: bücher
console.log("encode mañana:", encode("mañana"));
console.log("decode maana-pta:", decode("maana-pta"));
console.log("encode bücher:", encode("bücher"));
console.log("decode bcher-kva:", decode("bcher-kva"));

// @api: punycode.ucs2
// @expect: ucs2 decode length: 5
// @expect: ucs2 decode [0]: 72
// @expect: ucs2 decode [4]: 111
// @expect: ucs2 encode: Hello
const codePoints = ucs2.decode("Hello");
console.log("ucs2 decode length:", codePoints.length);
console.log("ucs2 decode [0]:", codePoints[0]);
console.log("ucs2 decode [4]:", codePoints[4]);
console.log("ucs2 encode:", ucs2.encode(codePoints));

// @api: punycode.version
// @expect: version: 2.1.0
// @expect: punycode toASCII: xn--maana-pta.com
// @expect: punycode.decode: mañana
console.log("version:", version);
console.log("punycode toASCII:", punycode.toASCII("mañana.com"));
console.log("punycode.decode:", punycode.decode("maana-pta"));
