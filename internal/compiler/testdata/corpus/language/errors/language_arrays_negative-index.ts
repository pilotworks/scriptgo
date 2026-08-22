// ScriptGo Corpus: Language Negative (language_arrays_negative-index)
// @run.err: array index must be a non-negative integer
const values: number[] = [10];
console.log(values[1 - 2]);
