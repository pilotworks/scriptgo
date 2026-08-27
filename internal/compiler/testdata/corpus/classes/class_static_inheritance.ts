// @expect: Base config: 42
// @expect: Sub config: 100
// @expect: Total instances: 2
class BaseCounter {
    static globalCount: number = 0;
    static getConfig(): number {
        return 42;
    }

    constructor() {
        BaseCounter.globalCount++;
    }
}

class SubCounter extends BaseCounter {
    static getSubConfig(): number {
        return 100;
    }
}

const c1 = new BaseCounter();
const c2 = new SubCounter();

console.log("Base config: " + BaseCounter.getConfig());
console.log("Sub config: " + SubCounter.getSubConfig());
console.log("Total instances: " + BaseCounter.globalCount);
