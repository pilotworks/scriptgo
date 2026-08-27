import { StringDecoder } from "node:string_decoder";
import { StringDecoder as StringDecoderBare } from "string_decoder";

// @api: StringDecoder.constructor
// @api: StringDecoder.encoding
// @expect: utf8
const decoder = new StringDecoder("utf8");
console.log(decoder.encoding);

// @api: StringDecoder.write
// @expect: part1: 
// @expect: part2: €
const euro = Buffer.from([0xE2, 0x82, 0xAC]); // € (3 bytes: E2 82 AC)
const part1 = decoder.write(euro.subarray(0, 2));
console.log("part1:", part1);
const part2 = decoder.write(euro.subarray(2));
console.log("part2:", part2);

// @expect: chunk1: 
// @expect: chunk2: 🚀 Launch!
const rocket = Buffer.from("🚀 Launch!");
const chunk1 = rocket.subarray(0, 2);
const chunk2 = rocket.subarray(2);
console.log("chunk1:", decoder.write(chunk1));
console.log("chunk2:", decoder.write(chunk2));

// @api: StringDecoder.end
// @expect: flushed length: true
const decoder2 = new StringDecoder("utf8");
const euroIncomplete = euro.subarray(0, 2);
decoder2.write(euroIncomplete);
const flushed = decoder2.end();
console.log("flushed length:", flushed.length > 0);

// @expect: b64Part1: 
// @expect: b64Part2: QUJD
// @expect: b64End: 
const b64Decoder = new StringDecoder("base64");
const b64Part1 = b64Decoder.write(Buffer.from([65, 66]));
console.log("b64Part1:", b64Part1);
const b64Part2 = b64Decoder.write(Buffer.from([67]));
console.log("b64Part2:", b64Part2);
const b64End = b64Decoder.end();
console.log("b64End:", b64End);

// @expect: hex: deadbeef
const hexDecoder = new StringDecoderBare("hex");
console.log("hex:", hexDecoder.write(Buffer.from([0xDE, 0xAD, 0xBE, 0xEF])));

// @expect: latin1: Hello
const latin1Decoder = new StringDecoder("latin1");
console.log("latin1:", latin1Decoder.write(Buffer.from([0x48, 0x65, 0x6C, 0x6C, 0x6F])));
