const buf = new ArrayBuffer(8);
const dv = new DataView(buf);
dv.setFloat64(0, 3.14);
console.log(dv.getFloat64(0));
