// @check.err: generators do not support a yield used as a value
// next(value) is not lowered, so a yield used as a value is rejected.
function* echo(): Generator<number, void, number> {
  const got = yield 1;
  console.log(got);
}
const g = echo();
g.next();
g.next(5);
