// ScriptGo Corpus: Language - errors (language_errors_array-index-type)
// @check.err: TypeScript type error
const values: number[] = [1, 2];
console.log(values['first']);
