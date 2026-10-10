// @expect: 12
// @expect: hi!
// @expect: 8
// @expect: 7
// @expect: 20
// A `this:` parameter only types the receiver; it is not a call argument.
function combine(this: void, a: number, b: number): number {
  return a * 10 + b;
}
console.log(combine(1, 2));

const shout = function (this: unknown, text: string): string {
  return text + "!";
};
console.log(shout("hi"));

class Counter {
  base = 5;
  add(this: Counter, amount: number): number {
    return this.base + amount;
  }
}
console.log(new Counter().add(3));

interface Adder {
  apply(this: Adder, value: number): number;
}
class PlusOne implements Adder {
  apply(this: Adder, value: number): number {
    return value + 1;
  }
}
const adder: Adder = new PlusOne();
console.log(adder.apply(6));

type Callback = (this: void, value: number) => number;
const double: Callback = (value) => value * 2;
console.log(double(10));
