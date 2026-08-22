// ScriptGo Corpus: Language - errors (language_errors_array-index-type)
// @check.err: TS7015
const values: number[] = [1, 2];
console.log(values['first']);
