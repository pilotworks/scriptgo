const encoder = new TextEncoder();
console.log(encoder.encoding);

const encoded = encoder.encode("Hello, ScriptGo! 🚀");
console.log(encoded.length);
console.log(encoded[0]);
console.log(encoded[7]);

const decoder = new TextDecoder();
console.log(decoder.encoding);
console.log(decoder.fatal);
console.log(decoder.ignoreBOM);

const decoded = decoder.decode(encoded);
console.log(decoded);

const dest = new Uint8Array(10);
const res = encoder.encodeInto("Hello World", dest);
console.log(res.read);
console.log(res.written);
console.log(decoder.decode(dest));

const empty = encoder.encode("");
console.log(empty.length);
console.log(decoder.decode(empty));
