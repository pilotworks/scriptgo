// Advanced Primitives: BigInt, Symbol, and RegExp literals

console.log("=== Advanced Primitives Demo ===");

// 1. BigInt
const bigA = 1000000000000000000n;
const bigB = 250000000000000000n;
const bigSum = bigA + bigB;
const bigMul = bigB * 4n;
console.log("BigInt sum: " + bigSum.toString());
console.log("BigInt multiply: " + bigMul.toString());

// 2. Symbol & Symbol Registry
const sym = Symbol("service.token");
const regSym = Symbol.for("app.global.key");
console.log("Symbol desc: " + sym.description);
console.log("Symbol keyFor: " + Symbol.keyFor(regSym));

// 3. RegExp literals
const pattern = /^[a-z]+$/i;
console.log("Regex test 'ScriptGo': " + pattern.test("ScriptGo"));
console.log("Regex test '12345': " + pattern.test("12345"));
