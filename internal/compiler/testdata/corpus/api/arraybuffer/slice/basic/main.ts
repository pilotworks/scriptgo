const buf = new ArrayBuffer(16);
const sliced = buf.slice(4, 12);
console.log(sliced.byteLength);
