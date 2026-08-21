const buf = new ArrayBuffer(4);
const dv = new DataView(buf);
dv.setInt32(0, 100000);
console.log(dv.getInt32(0));
