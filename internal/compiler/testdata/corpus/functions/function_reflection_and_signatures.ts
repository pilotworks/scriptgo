// @expect: 0
// @expect: 1
// @expect: 2
// @expect: 1
// @expect: 1
// @expect: zeroArgs
// @expect: singleArg
// @expect: defaultArg
function zeroArgs(): void {}
function singleArg(a: number): void {}
function twoArgs(a: number, b: string): void {}
function defaultArg(a: number, b: string = "default"): void {}
function restArg(a: number, ...rest: number[]): void {}

// 1. Function.prototype.length (arity before default/rest)
console.log(zeroArgs.length);
console.log(singleArg.length);
console.log(twoArgs.length);
console.log(defaultArg.length);
console.log(restArg.length);

// 2. Function.prototype.name
console.log(zeroArgs.name);
console.log(singleArg.name);
console.log(defaultArg.name);
