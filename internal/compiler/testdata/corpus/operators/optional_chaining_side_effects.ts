// @expect: undefined
// @expect: 0
// @expect: 20
// @expect: 1
// @expect: undefined
// @expect: 0
// @expect: 99
// @expect: 1

// Verifies that optional calls and optional index accesses short-circuit
// without evaluating side effects in arguments/indices, and evaluate to `undefined`.

type Caller = {
  calc?: (n: number) => number;
  items?: number[];
};

function testCall(caller: Caller | null | undefined): void {
  let sideEffects = 0;
  function getArg(): number {
    sideEffects++;
    return 10;
  }
  const res = caller?.calc?.(getArg());
  console.log(res);
  console.log(sideEffects);
}

function testIndex(caller: Caller | null | undefined): void {
  let sideEffects = 0;
  function getIndex(): number {
    sideEffects++;
    return 1;
  }
  const res = caller?.items?.[getIndex()];
  console.log(res);
  console.log(sideEffects);
}

testCall(null);

const validCaller: Caller = {
  calc: (n: number) => n * 2,
  items: [42, 99, 100],
};

testCall(validCaller);
testIndex(undefined);
testIndex(validCaller);
