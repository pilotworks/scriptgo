// Async Generators & Generator functions demo

function* subSequence(): Generator<number, void, unknown> {
  yield 100;
  yield 200;
}

function* mainSequence(): Generator<number, void, unknown> {
  yield 10;
  yield* subSequence();
  yield 20;
}

async function* asyncNumbers(): AsyncGenerator<number, void, unknown> {
  yield 1;
  yield 2;
  yield 3;
}

console.log("=== Generators Demo ===");
for (const n of mainSequence()) {
  console.log(`Sequence item: ${n}`);
}

console.log("=== Async Generators Demo ===");
for await (const x of asyncNumbers()) {
  console.log(`Async yielded: ${x}`);
}

export {};
