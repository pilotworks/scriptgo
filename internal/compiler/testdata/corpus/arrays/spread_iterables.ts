// @expect: zzz
// @expect: 7,7
// @expect: b,c
// @expect: 1,2,3
// @expect: 0,1,2,9
// Spreading an iterable that is not an array into an array literal runs the
// for..of iteration protocol.
function* repeat<T>(value: T, times: number): Generator<T> {
  for (let i = 0; i < times; i++) yield value;
}
console.log([...repeat("z", 3)].join(""));
console.log([...repeat(7, 2)].join(","));

function* letters(): Generator<string> {
  yield "a";
  yield "b";
  yield "c";
}
const it = letters();
it.next();
const rest: string[] = [...it];
console.log(rest.join(","));

const set = new Set<number>([1, 2, 2, 3]);
const fromSet: number[] = [...set];
console.log(fromSet.join(","));
const tuple: [number, number] = [1, 2];
const mixed: number[] = [0, ...tuple, ...new Set<number>([9])];
console.log(mixed.join(","));
