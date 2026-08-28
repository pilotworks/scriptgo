// @expect: step: 1, sum: 10
// @expect: step: 2, sum: 30
// @expect: step: 3, sum: 60
// @expect: final sum: 60

async function asyncMultiplier(val: number): Promise<number> {
  return val * 10;
}

async function runSequential(): Promise<number> {
  let sum = 0;
  for (let i = 1; i <= 3; i++) {
    const res = await asyncMultiplier(i);
    sum += res;
    console.log(`step: ${i}, sum: ${sum}`);
  }
  return sum;
}

async function main(): Promise<void> {
  const finalSum = await runSequential();
  console.log(`final sum: ${finalSum}`);
}

main();
