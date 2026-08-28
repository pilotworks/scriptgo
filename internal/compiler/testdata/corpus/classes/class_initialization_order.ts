// @expect: Base.field -> Base.constructor -> Middle.field -> Middle.constructor -> Derived.field -> Derived.constructor
const initLog: string[] = [];

function record(name: string): number {
    initLog.push(name);
    return 1;
}

class Base {
    baseField = record("Base.field");

    constructor() {
        initLog.push("Base.constructor");
    }
}

class Middle extends Base {
    middleField = record("Middle.field");

    constructor() {
        super();
        initLog.push("Middle.constructor");
    }
}

class Derived extends Middle {
    derivedField = record("Derived.field");

    constructor() {
        super();
        initLog.push("Derived.constructor");
    }
}

const instance = new Derived();
console.log(initLog.join(" -> "));
