// @run.err: Uncaught exception: TypeError: SG4002: cannot cast string to array
// Native difference: `as` on an unknown is a checked cast. A tag mismatch
// throws a built-in TypeError instance (uncaught here, so it ends the program).
const boxed: unknown = "s";
const values = boxed as number[];
console.log(values.length);
