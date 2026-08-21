const s1: symbol = Symbol("foo");
const s2: symbol = Symbol("foo");
console.log(s1 === s2);
console.log(s1 === s1);

const reg1: symbol = Symbol.for("app.key");
const reg2: symbol = Symbol.for("app.key");
console.log(reg1 === reg2);

const k1 = Symbol.keyFor(reg1);
console.log(k1);
const k2 = Symbol.keyFor(s1);
console.log(k2);

console.log(s1.description);
console.log(s1.toString());

const sEmpty: symbol = Symbol();
console.log(sEmpty.description);
console.log(sEmpty.toString());

const it1: symbol = Symbol.iterator;
const it2: symbol = Symbol.iterator;
console.log(it1 === it2);
console.log(it1.toString());

console.log(typeof s1);
