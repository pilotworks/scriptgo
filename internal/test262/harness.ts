// scriptgo test262 harness: the subset of harness/sta.js and harness/assert.js
// that the Static tier can compile. Assertions that need first-class
// constructors or property descriptors (assert.throws, verifyProperty) are
// intentionally absent, so tests using them report as unsupported.

class Test262Error extends Error {
  constructor(message?: string) {
    super(message === undefined ? "" : message);
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
}
