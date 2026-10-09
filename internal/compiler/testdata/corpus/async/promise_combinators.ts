// Promise.all, allSettled, any and race settle from every input, in input
// order, whatever the element types.

function delayN(ms: number, v: number): Promise<number> {
  return new Promise<number>((res) => setTimeout(() => res(v), ms));
}
function delayS(ms: number, v: string): Promise<string> {
  return new Promise<string>((res) => setTimeout(() => res(v), ms));
}

// @expect: 1,2
// @expect: a,b
// @expect: 5 X
// @expect: 1 3
// @expect: 0 0
async function all() {
  const ps = [Promise.resolve(1), Promise.resolve(2)];
  console.log((await Promise.all(ps)).join(","));
  const ss = [delayS(5, "a"), delayS(1, "b")];
  console.log((await Promise.all(ss)).join(","));
  const [n, s] = await Promise.all([delayN(10, 4), delayS(1, "x")]);
  console.log(n + 1, s.toUpperCase());
  const mixed = await Promise.all([1, delayN(2, 3)]);
  console.log(mixed[0], mixed[1]);
  const empty: Promise<number>[] = [];
  console.log((await Promise.all(empty)).length, (await Promise.allSettled(empty)).length);
}

// @expect: all rejected: first
async function allRejects() {
  try {
    await Promise.all([delayN(5, 1), Promise.reject(new Error("first"))]);
  } catch (e) {
    console.log("all rejected:", (e as Error).message);
  }
}

// @expect: ok 1
// @expect: rej boom
// @expect: ok 3
// @expect: 2 rejected
async function allSettled() {
  const ps = [Promise.resolve(1), Promise.reject(new Error("boom")), Promise.resolve(3)];
  for (const x of await Promise.allSettled(ps)) {
    if (x.status === "fulfilled") console.log("ok", x.value);
    else console.log("rej", (x.reason as Error).message);
  }
  const lit = await Promise.allSettled([Promise.resolve(1), Promise.reject(new Error("b2"))]);
  console.log(lit.length, lit[1].status);
}

// @expect: 2
// @expect: 5
async function any() {
  console.log(await Promise.any([Promise.reject(new Error("x")), Promise.resolve(2)]));
  const qs = [Promise.reject(new Error("x")), delayN(1, 5)];
  console.log(await Promise.any(qs));
}

// @expect: true 2 All promises were rejected
async function anyRejects() {
  try {
    await Promise.any([Promise.reject(new Error("e1")), Promise.reject(new Error("e2"))]);
  } catch (e) {
    console.log(e instanceof AggregateError, (e as AggregateError).errors.length, (e as Error).message);
  }
}

// @expect: 2
// @expect: late
async function race() {
  console.log(await Promise.race([delayN(20, 1), delayN(5, 2)]));
  const rs = [delayS(10, "late"), new Promise<string>(() => {})];
  console.log(await Promise.race(rs));
}

all().then(allRejects).then(allSettled).then(any).then(anyRejects).then(race);
