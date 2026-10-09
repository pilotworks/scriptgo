// @expect: 3 7
// @expect: true true
// @expect: true TypeError
// Destructuring performs RequireObjectCoercible on its source: null or
// undefined throws a TypeError before any binding is read, and enclosing
// finally blocks still run.
type Point = { x: number; y: number };
function sum(source: Point | undefined): number {
  const { x, y } = source as Point;
  return x + y;
}
function head(values: number[] | null): number {
  const [first] = values as number[];
  return first;
}
console.log(sum({ x: 1, y: 2 }), head([7, 8]));
let ran = false;
try {
  try {
    sum(undefined);
  } finally {
    ran = true;
  }
} catch (e) {
  console.log(ran, e instanceof TypeError);
}
try {
  head(null);
  console.log("not reached");
} catch (e) {
  console.log(e instanceof TypeError, (e as Error).name);
}
