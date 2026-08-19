class Counter {
    val: number;
    constructor(val: number) {
        this.val = val;
    }
    next(): number {
        this.val += 1;
        return 2;
    }
}

const c: Counter = new Counter(0);

// 1. Test single evaluation of switch expression
switch (c.next()) {
    case 1:
        console.log("one");
        break;
    case 2:
        console.log("two");
        break;
    default:
        console.log("other");
        break;
}
console.log(c.val); // Must be 1

// 2. Test fallthrough across non-empty statements
let tag: number = 1;
switch (tag) {
    case 1:
        console.log("start 1");
    case 2:
        console.log("fallthrough to 2");
        break;
    case 3:
        console.log("3");
        break;
    default:
        console.log("def");
        break;
}

// 3. Test break inside loop only breaks switch
let loopCount: number = 0;
for (let i: number = 0; i < 3; i += 1) {
    switch (i) {
        case 0:
            loopCount += 10;
            break; // should break switch, not for-loop!
        case 1:
            loopCount += 20;
            break;
        default:
            loopCount += 30;
            break;
    }
}
console.log(loopCount); // 10 + 20 + 30 = 60
