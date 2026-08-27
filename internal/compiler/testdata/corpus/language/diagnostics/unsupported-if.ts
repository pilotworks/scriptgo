// ScriptGo Corpus: Language - diagnostics (language_diagnostics_unsupported-if)
// @check.err: native subset
function check(condition: any) {
    if (condition) {
        console.log(condition);
    }
}
check(1);
