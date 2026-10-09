// ScriptGo Corpus: Language Negative (property on a function declaration)
// @run.err: property "count" on a function declaration is not supported
// Each reference to a function declaration is a fresh closure-ABI value, so
// properties assigned to it have nowhere to live.
function counter(): number {
  return 1;
}
// @ts-ignore TS2339: expando property assignment is the point of this case.
counter.count = 1;
console.log(counter());
