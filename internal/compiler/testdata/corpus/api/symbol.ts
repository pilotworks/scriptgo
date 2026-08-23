// ScriptGo Corpus: Symbol Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: symbol.for
// @expect: true
const s1_symbol_for_0: symbol = Symbol.for("foo");
const s2_symbol_for_0: symbol = Symbol.for("foo");
console.log(s1_symbol_for_0 === s2_symbol_for_0);

// @api: symbol.keyFor
// @expect: bar
const s_symbol_keyFor_2 = Symbol.for("bar");
console.log(Symbol.keyFor(s_symbol_keyFor_2));

// @api: symbol.symbol
// @api: Symbol.readonly description: string | undefined
// @expect: symbol
// @expect: test
const s_symbol_symbol_3 = Symbol("test");
console.log(typeof s_symbol_symbol_3);
console.log(s_symbol_symbol_3.description);

// @api: symbol.toString
// @expect: Symbol(test)
console.log(s_symbol_symbol_3.toString());

// @api: Symbol.valueOf(): symbol
// @expect: Symbol(test)
console.log(s_symbol_symbol_3.valueOf().toString());

// @api: symbol.iterator
// @api: Symbol.readonly asyncIterator: unique symbol
// @api: Symbol.readonly dispose: unique symbol
// @api: Symbol.readonly asyncDispose: unique symbol
// @api: Symbol.readonly hasInstance: unique symbol
// @api: Symbol.readonly isConcatSpreadable: unique symbol
// @api: Symbol.readonly match: unique symbol
// @api: Symbol.readonly matchAll: unique symbol
// @api: Symbol.readonly metadata: unique symbol
// @api: Symbol.readonly replace: unique symbol
// @api: Symbol.readonly search: unique symbol
// @api: Symbol.readonly species: unique symbol
// @api: Symbol.readonly split: unique symbol
// @api: Symbol.readonly toPrimitive: unique symbol
// @api: Symbol.readonly toStringTag: unique symbol
// @api: Symbol.readonly unscopables: unique symbol
// @expect: symbol
// @expect: symbol
// @expect: symbol
// @expect: symbol
// @expect: symbol
// @expect: symbol
// @expect: symbol
// @expect: symbol
// @expect: symbol
// @expect: symbol
// @expect: symbol
// @expect: symbol
// @expect: symbol
// @expect: symbol
// @expect: symbol
// @expect: symbol
console.log(typeof Symbol.iterator);
console.log(typeof Symbol.asyncIterator);
console.log(typeof Symbol.dispose);
console.log(typeof Symbol.asyncDispose);
console.log(typeof Symbol.hasInstance);
console.log(typeof Symbol.isConcatSpreadable);
console.log(typeof Symbol.match);
console.log(typeof Symbol.matchAll);
console.log(typeof Symbol.metadata);
console.log(typeof Symbol.replace);
console.log(typeof Symbol.search);
console.log(typeof Symbol.species);
console.log(typeof Symbol.split);
console.log(typeof Symbol.toPrimitive);
console.log(typeof Symbol.toStringTag);
console.log(typeof Symbol.unscopables);
