// @expect: 0,1,2,3,4
// @expect: created
// @expect: start
// @expect: a
// @expect: between
// @expect: b
// @expect: end
// @expect: true true
// @expect: r0c0 r0c2 r1c0 r1c2
// @expect: 1,2,0,1,2,10,20
// @expect: guarded 1
// @expect: caught boom
// @expect: guarded 2
// @expect: 0,10,11,20,21,22
// @expect: 3,2,1
// @expect: 0,1,1,2,3,5,8,13,21,34
// Generators run lazily: each next() resumes the body after its last yield,
// so infinite generators terminate when the consumer stops, side effects
// interleave with the consumer, and locals, loops, labels, try/catch and
// yield* keep their state between steps.
function* naturals(): Generator<number> {
  let n = 0;
  while (true) yield n++;
}
function* take(source: Generator<number>, count: number): Generator<number> {
  let taken = 0;
  for (const value of source) {
    if (taken >= count) return;
    taken++;
    yield value;
  }
}
console.log([...take(naturals(), 5)].join(","));

function* traced(): Generator<string> {
  console.log("start");
  yield "a";
  console.log("between");
  yield "b";
  console.log("end");
}
const t = traced();
console.log("created");
console.log(t.next().value);
console.log(t.next().value);
console.log(t.next().done, t.next().done);

function* grid(rows: number, cols: number): Generator<string> {
  for (let r = 0; r < rows; r++) {
    const label = "r" + r;
    for (let c = 0; c < cols; c++) {
      if (c === 1) continue;
      yield label + "c" + c;
    }
  }
}
console.log([...grid(2, 3)].join(" "));

function* delegating(): Generator<number> {
  yield* [1, 2];
  yield* take(naturals(), 3);
  const words = new Map<string, number>([["x", 10], ["y", 20]]);
  for (const [, v] of words) yield v;
}
console.log([...delegating()].join(","));

function* guarded(): Generator<number> {
  try {
    yield 1;
    throw new Error("boom");
  } catch (e) {
    console.log("caught", (e as Error).message);
  }
  yield 2;
}
for (const v of guarded()) console.log("guarded", v);

function* labelled(): Generator<number> {
  outer: for (let i = 0; i < 3; i++) {
    let j = 0;
    while (true) {
      if (j > i) continue outer;
      yield i * 10 + j;
      j++;
    }
  }
}
console.log([...labelled()].join(","));

const countdown = function* (from: number): Generator<number> {
  while (from > 0) yield from--;
};
console.log([...countdown(3)].join(","));

class Fib {
  *[Symbol.iterator](): Generator<number> {
    let a = 0;
    let b = 1;
    while (true) {
      yield a;
      [a, b] = [b, a + b];
    }
  }
}
const fib: number[] = [];
for (const f of new Fib()) {
  if (f > 50) break;
  fib.push(f);
}
console.log(fib.join(","));
