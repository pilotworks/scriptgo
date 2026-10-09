// @run.err: Uncaught exception: RangeError: boom
// A top-level throw ends main; an uncaught Error reports its name and message.
console.log("before");
throw new RangeError("boom");
