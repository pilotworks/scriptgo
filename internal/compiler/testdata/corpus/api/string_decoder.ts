// ScriptGo Corpus: Node.js String Decoder Module (Strict 1:1 Parity Tests)
import { StringDecoder } from "node:string_decoder";

// @api: new string_decoder.StringDecoder
// @expect: utf8
const decoder = new StringDecoder("utf8");
console.log(decoder.encoding);

// @api: StringDecoder.write
// @expect: true
const euro = Buffer.from([0xE2, 0x82, 0xAC]);
const part1 = decoder.write(euro.subarray(0, 2));
const part2 = decoder.write(euro.subarray(2));
console.log(part2 === "€");

// @api: StringDecoder.end
// @expect: true
const decoder2 = new StringDecoder("utf8");
decoder2.write(euro.subarray(0, 2));
const flushed = decoder2.end();
console.log(flushed.length > 0);
