// scriptgo test262 harness: the subset of harness/sta.js and harness/assert.js
// that the Static tier can compile, plus harness/compareArray.js. Constructors
// are not first-class values in the Static tier, so Prepare rewrites
// `assert.throws(TypeError, fn)` to `assert.throwsNamed("TypeError", fn)` for
// the built-in error constructors listed in errorName. Assertions that need
// property descriptors (verifyProperty) are absent and report as unsupported.

class Test262Error extends Error {
  constructor(message?: string) {
    super(message === undefined ? "" : message);
    this.name = "Test262Error";
  }
}

function $DONOTEVALUATE(): void {
  throw new Test262Error("Test262: This statement should not be evaluated.");
}

function assert(mustBeTrue: boolean, message?: string): void {
  if (mustBeTrue === true) {
    return;
  }
  throw new Test262Error(message === undefined ? "Expected true but got false" : message);
}

namespace assert {
  export function sameValue(actual: unknown, expected: unknown, message?: string): void {
    if (Object.is(actual, expected)) {
      return;
    }
    throw new Test262Error((message === undefined ? "" : message + " ") + "Expected SameValue(«" + String(actual) + "», «" + String(expected) + "») to be true");
  }

  export function notSameValue(actual: unknown, unexpected: unknown, message?: string): void {
    if (!Object.is(actual, unexpected)) {
      return;
    }
    throw new Test262Error((message === undefined ? "" : message + " ") + "Expected SameValue(«" + String(actual) + "», «" + String(unexpected) + "») to be false");
  }

  export function throwsNamed(expectedName: string, func: () => void, message?: string): void {
    const prefix = message === undefined ? "" : message + " ";
    try {
      func();
    } catch (thrown) {
      const actualName = errorName(thrown);
      if (actualName !== expectedName) {
        throw new Test262Error(prefix + "Expected a " + expectedName + " but got a " + actualName);
      }
      return;
    }
    throw new Test262Error(prefix + "Expected a " + expectedName + " to be thrown but no exception was thrown at all");
  }

  export function compareArray(actual: unknown[], expected: unknown[], message?: string): void {
    if (sameArrayContents(actual, expected)) {
      return;
    }
    const prefix = message === undefined ? "" : message + " ";
    throw new Test262Error(prefix + "Actual [" + actual.join(", ") + "] and expected [" + expected.join(", ") + "] should have the same contents.");
  }
}

// errorName names the constructor of a thrown value the way assert.throws
// compares constructors: most-derived built-in error first, typeof otherwise.
function errorName(value: unknown): string {
  if (value instanceof Test262Error) return "Test262Error";
  if (value instanceof TypeError) return "TypeError";
  if (value instanceof RangeError) return "RangeError";
  if (value instanceof SyntaxError) return "SyntaxError";
  if (value instanceof ReferenceError) return "ReferenceError";
  if (value instanceof URIError) return "URIError";
  if (value instanceof EvalError) return "EvalError";
  if (value instanceof Error) return "Error";
  return typeof value;
}

// compareArray is harness/compareArray.js: element-wise SameValue.
function compareArray(a: unknown[], b: unknown[]): boolean {
  return sameArrayContents(a, b);
}

function sameArrayContents(a: unknown[], b: unknown[]): boolean {
  if (b.length !== a.length) {
    return false;
  }
  for (let i = 0; i < a.length; i++) {
    if (!Object.is(b[i], a[i])) {
      return false;
    }
  }
  return true;
}
