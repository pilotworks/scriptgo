// Error.stackTraceLimit is a mutable setting that caps recorded stack frames.
function frames(e: Error): number {
  return (e.stack ?? "").split("\n").filter((l) => l.startsWith("    at ")).length;
}
function depth(n: number): Error { return n === 0 ? new Error("deep") : depth(n - 1); }
// @expect: 10
console.log(Error.stackTraceLimit);
Error.stackTraceLimit = 1;
// @expect: 1 true
console.log(Error.stackTraceLimit, frames(depth(5)) <= 1);
Error.stackTraceLimit = 0;
// @expect: 0 0 Error: deep
console.log(Error.stackTraceLimit, frames(depth(5)), (depth(1).stack ?? "").split("\n")[0]);
