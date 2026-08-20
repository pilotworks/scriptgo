const s1: symbol = Symbol.for("app.version");
const s2: symbol = Symbol.for("app.version");
const s3: symbol = Symbol.for("app.author");

console.log(s1 === s2);
console.log(s1 === s3);

const k1: string = Symbol.keyFor(s1)!;
const k3: string = Symbol.keyFor(s3)!;
console.log(k1);
console.log(k3);

const localSym1: symbol = Symbol("unique");
const localSym2: symbol = Symbol("unique");
console.log(localSym1 === localSym2);

console.log(s1.toString());
console.log(localSym1.toString());
