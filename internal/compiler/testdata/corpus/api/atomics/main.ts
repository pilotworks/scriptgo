const sab = new SharedArrayBuffer(1024);
console.log("SAB byteLength:", sab.byteLength);

const ta = new Int32Array(sab);
console.log("ta length:", ta.length);

Atomics.store(ta, 0, 100);
console.log("Loaded after store:", Atomics.load(ta, 0));

const oldAdd = Atomics.add(ta, 0, 25);
console.log("oldAdd:", oldAdd, "newVal:", Atomics.load(ta, 0));

const oldSub = Atomics.sub(ta, 0, 10);
console.log("oldSub:", oldSub, "newVal:", Atomics.load(ta, 0));

const oldEx = Atomics.exchange(ta, 0, 500);
console.log("oldEx:", oldEx, "newVal:", Atomics.load(ta, 0));

const oldCmp = Atomics.compareExchange(ta, 0, 500, 999);
console.log("oldCmp:", oldCmp, "newVal:", Atomics.load(ta, 0));

console.log("isLockFree 4:", Atomics.isLockFree(4));
console.log("isLockFree 3:", Atomics.isLockFree(3));
