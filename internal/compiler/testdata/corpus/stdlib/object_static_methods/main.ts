class Point {
  x: number;
  y: number;
  constructor(x: number, y: number) {
    this.x = x;
    this.y = y;
  }
}

class User {
  name: string;
  age: number;
  constructor(name: string, age: number) {
    this.name = name;
    this.age = age;
  }
}

function testObjectIs(): void {
  console.log("=== Object.is ===");
  console.log(Object.is(42, 42));
  console.log(Object.is(42, 43));
  console.log(Object.is("hello", "hello"));
  console.log(Object.is("hello", "world"));
  console.log(Object.is(true, true));
  console.log(Object.is(true, false));
  console.log(Object.is(NaN, NaN));
  console.log(Object.is(0, 0));
}

function testObjectHasOwn(): void {
  console.log("=== Object.hasOwn ===");
  const p = new Point(10, 20);
  console.log(Object.hasOwn(p, "x"));
  console.log(Object.hasOwn(p, "y"));
  console.log(Object.hasOwn(p, "z"));
}

function testObjectKeys(): void {
  console.log("=== Object.keys ===");
  const p = new Point(1, 2);
  const keys: string[] = Object.keys(p);
  for (const k of keys) {
    console.log(k);
  }
}

function testObjectValues(): void {
  console.log("=== Object.values ===");
  const p = new Point(100, 200);
  const vals: number[] = Object.values(p);
  for (const v of vals) {
    console.log(v);
  }
}

function testObjectAssign(): void {
  console.log("=== Object.assign ===");
  const p1 = new Point(1, 2);
  const p2 = new Point(10, 20);
  const res: Point = Object.assign(p1, p2);
  console.log(res.x);
  console.log(res.y);
  console.log(p1.x);
  console.log(p1.y);
}

function main(): void {
  testObjectIs();
  testObjectHasOwn();
  testObjectKeys();
  testObjectValues();
  testObjectAssign();
}

main();
