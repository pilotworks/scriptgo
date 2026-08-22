// ScriptGo Corpus: Symbol Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: symbol.for
// @expect: true
const s1_symbol_for_0: symbol = Symbol.for("foo");
const s2_symbol_for_0: symbol = Symbol.for("foo");
console.log(s1_symbol_for_0 === s2_symbol_for_0);

// @api: symbol.iterator
// @expect: symbol
// @expect: symbol
console.log(typeof Symbol.iterator); console.log(typeof Symbol.asyncIterator);

// @api: symbol.keyFor
// @expect: bar
const s_symbol_keyFor_2 = Symbol.for("bar"); console.log(Symbol.keyFor(s_symbol_keyFor_2));

// @api: symbol.symbol
// @expect: symbol
const s_symbol_symbol_3 = Symbol("test"); console.log(typeof s_symbol_symbol_3);

// @api: symbol.toString
// @expect: Symbol(test)
const s_symbol_toString_4 = Symbol("test");
console.log(s_symbol_toString_4.toString());
