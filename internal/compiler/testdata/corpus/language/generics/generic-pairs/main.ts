function swap<T, U>(first: T, second: U): [U, T] {
  return [second, first];
}

function mapPair<T, U>(first: T, second: T, fn: (val: T) => U): [U, U] {
  return [fn(first), fn(second)];
}

const p1 = swap<number, string>(42, "hello");
console.log(p1[0]);
console.log(p1[1]);

const p2 = swap<boolean, number>(true, 100);
console.log(p2[0]);
console.log(p2[1]);

const doubleNum = (x: number): number => x * 2;
const mappedNums = mapPair<number, number>(5, 10, doubleNum);
console.log(mappedNums[0]);
console.log(mappedNums[1]);

const strLen = (s: string): number => s.length;
const mappedStrings = mapPair<string, number>("apple", "banana", strLen);
console.log(mappedStrings[0]);
console.log(mappedStrings[1]);
