const buf = new ArrayBuffer(4);
const dv = new DataView(buf);
dv.setInt8(0, 42);
console.log(dv.getInt8(0));
