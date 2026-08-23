// @expect: 3
// @expect: 20
// @expect: 40
class Counter {
    val: number = 0;
    inc(): number {
        this.val++;
        return this.val;
    }
}

const c = new Counter();
const res = (c.inc(), c.inc(), c.inc());
console.log(res);

let a = 10;
let b = (a = 20, a * 2);
console.log(a);
console.log(b);
