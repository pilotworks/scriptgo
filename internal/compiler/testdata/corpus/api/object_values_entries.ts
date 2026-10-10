// Object.values/entries list the fields an object has (absent optional
// fields are skipped) with each value's own type.
// @expect: 1 ["x"]
// @expect: [1,"y"] [["a",1],["b","y"]]
// @expect: [2,"z",true]
type Opt = { a?: number; b: string };
const o: Opt = { b: "x" };
console.log(Object.values(o).length, JSON.stringify(Object.values(o)));
const p: Opt = { a: 1, b: "y" };
console.log(JSON.stringify(Object.values(p)), JSON.stringify(Object.entries(p)));
type Mixed = { n: number; s: string; f: boolean };
const m: Mixed = { n: 2, s: "z", f: true };
console.log(JSON.stringify(Object.values(m)));

// @expect: 3 2
// @expect: x 10
// @expect: y 20
// @expect: 8
// @expect: AMY 4
// @expect: BOB 6
// @expect: [["q","w"]]
// @expect: [ false, 0 ]
class Pt { x = 1; y = 2; }
const pt = new Pt();
const nums: number[] = Object.values(pt);
console.log(nums.reduce((a, b) => a + b, 0), nums.length);
const pairs: [string, number][] = Object.entries(pt);
for (const [k, v] of pairs) console.log(k, v * 10);
const scores: Record<string, number> = { amy: 3, bob: 5 };
let total = 0;
for (const v of Object.values(scores)) total += v;
console.log(total);
for (const [name, score] of Object.entries(scores)) console.log(name.toUpperCase(), score + 1);
console.log(JSON.stringify(Object.entries({ q: "w" })));
console.log(Object.values({ flag: false, n: 0 }));
