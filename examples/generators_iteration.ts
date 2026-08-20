// Synchronous and asynchronous generators with for..of iteration

function* idGenerator(prefix: string, count: number): Generator<string, void, unknown> {
  let index = 1;
  while (index <= count) {
    yield `${prefix}-${index}`;
    index++;
  }
}

async function* asyncDataStream(): AsyncGenerator<number, void, unknown> {
  const values = [10, 20, 30, 40];
  for (const v of values) {
    yield v * 2;
  }
}

console.log("=== Synchronous Generator ===");
const gen = idGenerator("JOB", 3);
for (const id of gen) {
  console.log("Generated ID:", id);
}

console.log("=== Asynchronous Generator Stream ===");
async function consumeAsyncStream(): Promise<void> {
  for await (const val of asyncDataStream()) {
    console.log("Streamed Value:", val);
  }
}

consumeAsyncStream();
