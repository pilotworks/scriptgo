// @check.err: generators do not support a yield inside a catch block or a try statement with finally
// Suspending inside try/finally would run the finally block on every yield.
function* cleanup(): Generator<number> {
  try {
    yield 1;
  } finally {
    console.log("done");
  }
}
for (const v of cleanup()) console.log(v);
