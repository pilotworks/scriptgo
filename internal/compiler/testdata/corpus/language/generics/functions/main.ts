function identity<T>(value: T): T {
  return value;
}

function pickFirst<T, U>(first: T, second: U): T {
  return first;
}

function pickSecond<T, U>(first: T, second: U): U {
  return second;
}

function wrapInArray<T>(item: T): T[] {
  return [item];
}

function getFirst<T>(arr: T[]): T {
  return arr[0];
}

function getLast<T>(arr: T[]): T {
  return arr[arr.length - 1];
}

const n = identity<number>(42);
const s = identity<string>("hello generics");
const inferredNum = identity(99);

console.log(n);
console.log(s);
console.log(inferredNum);
console.log(pickFirst<number, string>(123, "ignored"));
console.log(pickSecond<string, number>("ignored", 456));
console.log(wrapInArray<number>(555)[0]);
console.log(getFirst<number>([100, 200, 300]));
console.log(getLast<string>(["first", "second", "third"]));
