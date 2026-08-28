// @expect: A,B,C
// @expect: 20
// @expect: X,Y,Z,W
// @expect: true
// @expect: F1
// @expect: false
// @expect: arg1,arg2,arg3
// @expect: 60
// @expect: n1,n2,n3
// @expect: 123
let log: string[] = [];

function track(tag: string, val: number): number {
    log.push(tag);
    return val;
}

// 1. Binary arithmetic: (a + b) * c
log = [];
const r1 = (track("A", 2) + track("B", 3)) * track("C", 4);
console.log(log.join(","));
console.log(r1);

// 2. Binary comparison: a < b && c > d
log = [];
const r2 = track("X", 5) < track("Y", 10) && track("Z", 20) > track("W", 15);
console.log(log.join(","));
console.log(r2);

// 3. Short circuit evaluation
log = [];
const r3 = track("F1", 0) > 10 && track("F2", 5) > 0;
console.log(log.join(","));
console.log(r3);

// 4. Function call arguments left to right
log = [];
function add3(a: number, b: number, c: number): number {
    return a + b + c;
}
const r4 = add3(track("arg1", 10), track("arg2", 20), track("arg3", 30));
console.log(log.join(","));
console.log(r4);

// 5. Nested calls evaluation order
log = [];
function combine(x: number, y: number): number {
    return x * 10 + y;
}
const r5 = combine(combine(track("n1", 1), track("n2", 2)), track("n3", 3));
console.log(log.join(","));
console.log(r5);
