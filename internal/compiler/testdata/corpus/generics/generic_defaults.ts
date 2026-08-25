// @expect: default string
// @expect: 12345
class Box<T = string> {
    value: T;
    constructor(val: T) {
        this.value = val;
    }
    getValue(): T {
        return this.value;
    }
}

const b1 = new Box("default string");
console.log(b1.getValue());

const b2 = new Box<number>(12345);
console.log(b2.getValue());
