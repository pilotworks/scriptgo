// @expect: ok
// @expect: f
// @expect: after
// @expect: caught z
// An empty catch clause still catches; an empty finally clause without a
// catch still rethrows.
try {
  throw new Error("x");
} catch {}
console.log("ok");
try {
  throw new Error("y");
} catch {
} finally {
  console.log("f");
}
console.log("after");
try {
  try {
    throw new Error("z");
  } finally {
  }
} catch (e) {
  console.log("caught", (e as Error).message);
}
