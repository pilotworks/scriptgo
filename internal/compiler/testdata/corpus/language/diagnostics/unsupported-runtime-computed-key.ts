// ScriptGo Corpus: Language Negative (runtime computed property key)
// @check.err: SG2005
// A computed key known only at run time cannot be part of a fixed shape.
const suffix: string = String(Math.random() > 2);
const record = { ["k" + suffix]: 1 };
console.log(record);
