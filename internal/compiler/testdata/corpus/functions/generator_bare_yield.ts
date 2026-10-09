// @expect: 1
// @expect: undefined
// @expect: 3
// A bare `yield;` produces undefined.
function* steps(): Generator<number | undefined> {
  yield 1;
  yield;
  yield 3;
}
for (const step of steps()) {
  console.log(step);
}
