interface Container<T> {
    value: T;
}

class Box<T> {
    item: T;
    constructor(item: T) {
        this.item = item;
    }
    getItem(): T {
        return this.item;
    }
}

function wrap<T>(val: T): Container<T> {
    const c: Container<T> = { value: val };
    return c;
}

const numBox = new Box<number>(42);
console.log(numBox.getItem());

const strBox = new Box<string>("ScriptGo");
console.log(strBox.getItem());

const wrapped = wrap<string>("boxed");
console.log(wrapped.value);
