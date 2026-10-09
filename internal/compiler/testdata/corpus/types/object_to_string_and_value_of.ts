// toString/valueOf use an object's own methods (object literal properties or
// class methods) in direct calls, template literals and String().
// @expect: O(x) 42
// @expect: O(x) O(x)
// @expect: $2.50 $2.50 $2.50 250
// @expect: [object Object] [object Object]
const o = { name: "x", toString() { return "O(x)"; }, valueOf() { return 42; } };
console.log(o.toString(), o.valueOf());
console.log(`${o}`, String(o));
class Money { constructor(public cents: number) {} toString() { return "$" + (this.cents / 100).toFixed(2); } valueOf() { return this.cents; } }
const m = new Money(250);
console.log(m.toString(), `${m}`, String(m), m.valueOf());
const plain = { a: 1 };
console.log(plain.toString(), String(plain));

// @expect: [object Map] [object Set] [object Object]
console.log(new Map<string, number>().toString(), new Set<number>().toString(), ({}).toString());
