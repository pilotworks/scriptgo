// ScriptGo Corpus: Language - diagnostics (language_diagnostics_unsupported-while)
// @check.err: native subset
function check(condition: any) {
    while (condition) {
        console.log(condition);
    }
}
check(1);
