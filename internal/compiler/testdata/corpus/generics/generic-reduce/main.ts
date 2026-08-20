function customReduce<T, R>(arr: T[], initial: R, reducer: (acc: R, item: T) => R): R {
  let acc: R = initial;
  for (let i = 0; i < arr.length; i = i + 1) {
    acc = reducer(acc, arr[i]);
  }
  return acc;
}

const numbers: number[] = [1, 2, 3, 4, 5];
const sum = customReduce<number, number>(numbers, 0, (acc: number, x: number) => acc + x);
console.log(sum);

const product = customReduce<number, number>(numbers, 1, (acc: number, x: number) => acc * x);
console.log(product);

const words: string[] = ["a", "b", "c", "d"];
const joined = customReduce<string, string>(words, "start:", (acc: string, s: string) => acc + s);
console.log(joined);

const totalLength = customReduce<string, number>(words, 0, (acc: number, s: string) => acc + s.length);
console.log(totalLength);
