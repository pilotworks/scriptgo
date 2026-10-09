// @expect: Error name must not be empty true
// @expect: missing id true true
// @expect: bad type TypeError true true false
// @expect: TypeError: bad type TypeError: plain
// @expect: Error: a | RangeError: b
// User classes extending Error/native errors inherit message/name/stack, instanceof follows the native chain, and String() uses Error.prototype.toString.
class ValidationError extends Error {
  field: string;
  constructor(field: string, message: string) {
    super(message);
    this.field = field;
  }
}
class NotFound extends ValidationError {}
class Implicit extends TypeError {}
function check(value: string): void {
  if (value === "") {
    throw new ValidationError("name", "must not be empty");
  }
}
try {
  check("");
} catch (error) {
  if (error instanceof ValidationError) {
    console.log(error.name, error.field, error.message, error instanceof Error);
  }
}
const nf = new NotFound("id", "missing");
console.log(nf.message, nf.field, nf instanceof ValidationError, nf instanceof Error);
const im = new Implicit("bad type");
console.log(im.message, im.name, im instanceof TypeError, im instanceof Error, im instanceof RangeError);
console.log(String(im), `${new TypeError("plain")}`);
const errs: Error[] = [new Error("a"), new RangeError("b")];
console.log(errs.join(" | "));
