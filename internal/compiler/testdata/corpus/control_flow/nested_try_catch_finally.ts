// @expect: try1 try2 catch2 finally2 catch1 finally1
function complexTry(): string {
    let trace = "";
    try {
        trace += "try1 ";
        try {
            trace += "try2 ";
            throw new Error("err");
        } catch (e) {
            trace += "catch2 ";
            throw e;
        } finally {
            trace += "finally2 ";
        }
    } catch (e) {
        trace += "catch1 ";
    } finally {
        trace += "finally1";
    }
    return trace;
}

console.log(complexTry());
