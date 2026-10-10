// @expect: 3
// @expect: done
// @expect: 1
// @expect: 3
// @expect: 0
// @expect: 1
// @expect: 5
// @expect: fin
// @expect: 1104
// @expect: finally 1
// @expect: caught inner
// @expect: 5
// @expect: 10
// @expect: 9
// @expect: 5 6 7
// @expect: swallowed
// An async function may run several suspending units in sequence: awaits
// before and after a try, a loop, or a branch that also awaits.
async function num(v: number): Promise<number> {
  return v;
}
async function fail(msg: string): Promise<number> {
  throw new Error(msg);
}

async function awaitThenTry(): Promise<void> {
  const a = await num(1);
  try {
    const b = await num(2);
    console.log(a + b);
  } catch (e) {
    console.log("err");
  }
  console.log("done");
}

async function tryThenAwait(): Promise<void> {
  try {
    console.log(await num(1));
  } catch (e) {
    console.log("err");
  }
  const c = await num(3);
  console.log(c);
}

async function loopThenTry(): Promise<void> {
  for (let i = 0; i < 2; i++) console.log(await num(i));
  try {
    console.log(await num(5));
  } finally {
    console.log("fin");
  }
}

async function sum(limit: number): Promise<number> {
  let total = await num(1);
  for (let i = 0; i < limit; i++) total += await num(i);
  try {
    total += await fail("boom");
  } catch (e) {
    total += 100;
  }
  if (total > 50) {
    return total + (await num(1000));
  }
  return -1;
}

// A try without a catch clause propagates the rejection after finally.
async function rethrow(): Promise<string> {
  const a = await num(1);
  try {
    await fail("inner");
  } finally {
    console.log("finally", a);
  }
  return "unreachable";
}

class Counter {
  count = 0;
  async bump(times: number): Promise<number> {
    this.count += await num(1);
    for (let i = 0; i < times; i++) this.count += await num(2);
    return this.count;
  }
}

async function main(): Promise<void> {
  await awaitThenTry();
  await tryThenAwait();
  await loopThenTry();
  console.log(await sum(3));
  try {
    await rethrow();
  } catch (e) {
    console.log("caught", (e as Error).message);
  }
  console.log(await new Counter().bump(2));
  const base = 7;
  const add = async (x: number): Promise<number> => {
    const y = await num(x);
    try {
      return base + y + (await num(1));
    } catch (e) {
      return 0;
    }
  };
  console.log(await add(2));
  const plain = async (x: number): Promise<number> => {
    const y = await num(x);
    return base + y;
  };
  console.log(await plain(2));
  console.log(await num(5), await num(6), await num(7));
  try {
    await fail("quiet");
  } catch {}
  console.log("swallowed");
}
main();
