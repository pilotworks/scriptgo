// AggregateError carries its errors array in the Error layout.

// @expect: msg AggregateError 2 true true
const e = new AggregateError([new Error("a"), 2], "msg");
console.log(e.message, e.name, e.errors.length, e instanceof AggregateError, e instanceof Error);

// @expect: a
console.log((e.errors[0] as Error).message);
