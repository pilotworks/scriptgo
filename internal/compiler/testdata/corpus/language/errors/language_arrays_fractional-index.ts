// ScriptGo Corpus: Language Negative (language_arrays_fractional-index)
// @run.err: array index must be a non-negative integer
const values: number[] = [10];
console.log(values[0.5]);
