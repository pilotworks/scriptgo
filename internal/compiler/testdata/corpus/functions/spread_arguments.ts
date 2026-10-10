// @expect: 6
// @expect: 16
// @expect: 112
// @expect: p:a,b
// @expect: 3
// @expect: 4
// @expect: acde
// @expect: sum=20
// @expect: 4
// @expect: 0
// @expect: 3
// @expect: number,string,boolean
// @expect: 4|5|6
// @expect: 2
// @expect: 1a
// @expect: 2b
// @expect: 2bc
// @expect: 3 1 3
// @expect: [child-a-b-x-y]
// @expect: 1
// @expect: [child+x]
// @expect: 0 0 2 0
// Spread arguments fill a rest parameter (gathered with any other trailing
// arguments into one array); a fixed-length tuple variable spreads into
// ordinary parameters.
function sum(...xs: number[]): number {
  let total = 0;
  for (const x of xs) total += x;
  return total;
}
const arr = [1, 2, 3];
console.log(sum(...arr));
console.log(sum(10, ...arr));
console.log(sum(...arr, 100, ...arr));

function tag(prefix: string, ...rest: string[]): string {
  return prefix + ":" + rest.join(",");
}
const words = ["a", "b"];
console.log(tag("p", ...words));
console.log(Math.max(...arr));

class Point {
  x: number;
  y: number;
  constructor(x: number, y: number) {
    this.x = x;
    this.y = y;
  }
}
class Bag {
  items: string[] = [];
  add(...xs: string[]): number {
    this.items.push(...xs);
    return this.items.length;
  }
  static total(label: string, ...points: Point[]): string {
    let t = 0;
    for (const p of points) t += p.x + p.y;
    return label + "=" + t;
  }
}
const bag = new Bag();
const more = ["c", "d"];
console.log(bag.add("a", ...more, "e"));
console.log(bag.items.join(""));
const points = [new Point(1, 2), new Point(3, 4)];
console.log(Bag.total("sum", ...points, new Point(5, 5)));

function count(...values: unknown[]): number {
  return values.length;
}
const mixed: unknown[] = [1, "x", true];
console.log(count(...mixed, 4));
const empty: number[] = [];
console.log(count(...empty));
console.log(sum(...[7, 8], ...empty, 9) === 24 ? 3 : -1);

function kinds(...values: unknown[]): string {
  return values.map((v) => typeof v).join(",");
}
const tuple: [number, string] = [1, "a"];
console.log(kinds(...tuple, true));

class List {
  items: number[];
  constructor(...xs: number[]) {
    this.items = xs;
  }
}
const nums = [4, 5];
console.log(new List(...nums, 6).items.join("|"));
const length = (...xs: number[]): number => xs.length;
console.log(length(...nums));

function pair(a: number, b: string): string {
  return a + b;
}
console.log(pair(...tuple));
type Labeled = [x: number, label: string];
const labeled: Labeled = [2, "b"];
console.log(pair(...labeled));
function lead(a: number, ...rest: string[]): string {
  return a + rest.join("");
}
console.log(lead(...labeled, "c"));

const spreadTuple: number[] = [...([1, 2] as [number, number]), 3];
console.log(spreadTuple.length, spreadTuple[0], spreadTuple[2]);

class Base {
  parts: string[];
  constructor(name: string, ...parts: string[]) {
    this.parts = [name, ...parts];
  }
  join(sep: string, ...extra: string[]): string {
    return this.parts.concat(extra).join(sep);
  }
}
class Child extends Base {
  constructor(...parts: string[]) {
    super("child", ...parts);
  }
  join(sep: string, ...extra: string[]): string {
    return "[" + super.join(sep, "x", ...extra) + "]";
  }
}
console.log(new Child("a", "b").join("-", "y"));
console.log(new Base("solo").parts.length);
console.log(new Child().join("+"));

// An omitted optional parameter before a rest parameter leaves the rest empty.
function optionalThenRest(first?: number, ...rest: number[]): number {
  return rest.length;
}
class Statement {
  all(named?: string | number, ...anonymous: (string | number)[]): number {
    return anonymous.length;
  }
  iterate(named?: string | number, ...anonymous: (string | number)[]): number {
    return this.all(named, ...anonymous);
  }
  static count(named?: string, ...anonymous: string[]): number {
    return anonymous.length;
  }
}
const statement = new Statement();
console.log(optionalThenRest(), statement.iterate(), statement.iterate(1, "a", 2), Statement.count());
