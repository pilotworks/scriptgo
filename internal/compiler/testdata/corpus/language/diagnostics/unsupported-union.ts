// ScriptGo Corpus: Language - diagnostics (language_diagnostics_unsupported-union)
// @check.err: SG1002
function printVal(val: string | number) {
    const x = +val;
    console.log(x);
}
printVal("test");
