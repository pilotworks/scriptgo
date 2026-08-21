const encoder = new TextEncoder();
const u8: Uint8Array = encoder.encode("hi");
console.log(u8.length);
console.log(u8[0]);
