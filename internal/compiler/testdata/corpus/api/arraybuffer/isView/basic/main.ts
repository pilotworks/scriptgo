const u8 = new Uint8Array([1, 2, 3]);
console.log(ArrayBuffer.isView(u8));
console.log(ArrayBuffer.isView("not a view"));
