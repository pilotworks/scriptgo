const i32 = new Int32Array(3);
i32[0] = 100000;
i32[1] = -50000;
i32[2] = 42;

console.log(i32.length);
console.log(i32.byteLength);
console.log(i32.BYTES_PER_ELEMENT);
console.log(i32[0]);
console.log(i32[1]);
console.log(i32[2]);
