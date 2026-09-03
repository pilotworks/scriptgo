// ScriptGo Corpus: indexed assignment
// @run.expected: uppercase
const values: string[] = ["UPPERCASE"];
values[0] = values[0].toLowerCase();
console.log(values[0]);
