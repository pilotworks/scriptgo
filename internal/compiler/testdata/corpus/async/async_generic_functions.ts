// @expect: 6
// @expect: x!
// @expect: true
// @expect: b
// @expect: 3 undefined
// @expect: 1,2,3
// Generic async and async generator functions are specialized per call like
// other generic functions.
async function delay<T>(value: T): Promise<T> {
  await Promise.resolve();
  return value;
}
async function firstOf<T>(items: T[]): Promise<T | undefined> {
  const copy = await delay(items);
  return copy.length > 0 ? copy[0] : undefined;
}
async function* countUp<T>(values: T[]): AsyncGenerator<T> {
  for (const v of values) yield await delay(v);
}
class Box {
  label: string;
  constructor(label: string) {
    this.label = label;
  }
}
async function main(): Promise<void> {
  console.log((await delay(5)) + 1);
  console.log((await delay("x")) + "!");
  console.log(await delay(true));
  console.log((await delay(new Box("b"))).label);
  const first = await firstOf([3, 4]);
  const none = await firstOf<string>([]);
  console.log(first, none);
  const seen: number[] = [];
  for await (const v of countUp([1, 2, 3])) seen.push(v);
  console.log(seen.join(","));
}
main();
