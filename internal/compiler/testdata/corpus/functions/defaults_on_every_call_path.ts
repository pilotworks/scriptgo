// @expect: 30 6 20
// @expect: 3 4
// @expect: hello world hello ts hello world|hello world
// @expect: 42 2,4,6 1px 1em
// @expect: 2
// Parameter defaults apply on every call path: direct calls, closures,
// functions and static methods used as values (closure ABI trampolines).
const scale = (value: number = 5, factor: number = value + 1): number => value * factor;
console.log(scale(), scale(2), scale(2, 10));
const pick = function ({ x }: { x: number } = { x: 3 }): number {
  return x;
};
console.log(pick(), pick({ x: 4 }));
function greet(name: string = "world"): string {
  return "hello " + name;
}
const greetRef = greet;
console.log(greetRef(), greetRef("ts"), ["a", "b"].map(() => greetRef()).join("|"));
class Units {
  static double(n: number): number {
    return n * 2;
  }
  static label(unit: string = "px"): string {
    return "1" + unit;
  }
}
const double = Units.double;
const label = Units.label;
console.log(double(21), [1, 2, 3].map(Units.double).join(), label(), label("em"));
let evaluated = 0;
const counted = (n: number = ++evaluated): number => n;
counted();
counted(9);
counted();
console.log(evaluated);
