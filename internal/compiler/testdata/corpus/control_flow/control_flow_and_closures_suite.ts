// @expect: low_priority
// @expect: low_priority
// @expect: priority
// @expect: critical
// @expect: unknown
// @expect: 0-5-10-15
// @expect: 7 3
// @expect: 0:10,2:8,4:6
// 1. Complex switch-case fallthrough
function parseSeverity(level: number): string {
    let output = "";
    switch (level) {
        case 1:
        case 2:
            output += "low_";
        case 3:
            output += "priority";
            break;
        case 4:
            output += "critical";
            break;
        default:
            output += "unknown";
    }
    return output;
}

console.log(parseSeverity(1));
console.log(parseSeverity(2));
console.log(parseSeverity(3));
console.log(parseSeverity(4));
console.log(parseSeverity(99));

// 2. Closure loop capture per iteration
function createMultipliers(): number[] {
    const fnList: (() => number)[] = [];
    for (let idx = 0; idx < 4; idx++) {
        fnList.push(() => idx * 5);
    }
    return fnList.map(fn => fn());
}

const evaluated: number[] = createMultipliers();
console.log(evaluated.join("-"));

// 3. Nested labeled while loops with continue to outer label
let totalCount: number = 0;
let outerIdx: number = 0;

outerLoop: while (outerIdx < 3) {
    outerIdx++;
    let innerIdx: number = 0;
    while (innerIdx < 3) {
        innerIdx++;
        if (outerIdx === 2 && innerIdx === 2) {
            continue outerLoop;
        }
        totalCount++;
    }
}
console.log(totalCount, outerIdx);

// 4. Comma operator in for loop initialization and stepping
const stepping: string[] = [];
for (let p = 0, q = 10; p < q; p += 2, q -= 2) {
    stepping.push(p + ":" + q);
}
console.log(stepping.join(","));
