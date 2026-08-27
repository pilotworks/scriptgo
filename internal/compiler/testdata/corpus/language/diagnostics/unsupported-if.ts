// ScriptGo Corpus: Language - diagnostics (language_diagnostics_unsupported-if)
// @check.err: native subset
function check(condition: number | string) {
    if (condition) {
        console.log(condition);
    }
}
check(1);
