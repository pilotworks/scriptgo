const buf: Buffer = Buffer.alloc(16);

buf.writeUInt8(250, 0);
console.log(buf.readUInt8(0));

buf.writeInt8(-10, 1);
console.log(buf.readInt8(1));

buf.writeUInt16LE(4660, 2);
console.log(buf.readUInt16LE(2));
console.log(buf.readUInt16BE(2));

buf.writeUInt32LE(305419896, 4);
console.log(buf.readUInt32LE(4));

buf.writeInt32BE(-123456, 8);
console.log(buf.readInt32BE(8));
