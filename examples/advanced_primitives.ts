// Advanced Primitives: BigInt, Symbol, and RegExp literals

console.log("=== Advanced Primitives Demo ===");

// 1. BigInt
const bigA: bigint = 1000000000000000000n;
const bigB: bigint = 250000000000000000n;
const bigSum: bigint = bigA + bigB;
const bigMul: bigint = bigB * 4n;
console.log(`BigInt sum: ${bigSum}`);
console.log(`BigInt multiply: ${bigMul}`);

// 2. Symbol & Symbol Registry
const sym: symbol = Symbol("service.token");
const regSym: symbol = Symbol.for("app.global.key");
console.log(`Symbol desc: ${sym.description ?? ""}`);
console.log(`Symbol keyFor: ${Symbol.keyFor(regSym) ?? ""}`);

// 3. RegExp literals
const pattern: RegExp = /^[a-z]+$/i;
console.log(`Regex test 'ScriptGo': ${pattern.test("ScriptGo")}`);
console.log(`Regex test '12345': ${pattern.test("12345")}`);
