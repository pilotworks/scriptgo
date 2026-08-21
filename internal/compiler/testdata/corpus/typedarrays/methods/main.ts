const src = new Uint8Array(6);
src.fill(7, 1, 4);
console.log(src[0]);
console.log(src[1]);
console.log(src[2]);
console.log(src[3]);
console.log(src[4]);

const sub = src.subarray(1, 4);
console.log(sub.length);
console.log(sub.byteOffset);
console.log(sub[0]);

sub[0] = 42;
console.log(src[1]); // verifies subarray shares underlying buffer

const sl = src.slice(1, 4);
console.log(sl.length);
console.log(sl[0]);
sl[0] = 99;
console.log(src[1]); // verifies slice copies buffer (src[1] is still 42)

const dest = new Uint8Array(8);
dest.set(sub, 2);
console.log(dest[2]);
console.log(dest[3]);
