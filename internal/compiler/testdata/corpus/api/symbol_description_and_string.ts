// A symbol's description is undefined when none is given and is kept as
// given otherwise (including ""); String(symbol) is "Symbol(desc)".

// @expect: undefined true tag Symbol() Symbol() Symbol(tag) true
const s = Symbol();
const e = Symbol("");
const t = Symbol("tag");
console.log(s.description, e.description === "", t.description, String(s), String(e), t.toString(), s.description === undefined);

// @expect: true Symbol()
const u = Symbol(undefined);
console.log(u.description === undefined, String(u));
