// @expect: 100
// @expect: 200
// @expect: 60
// @expect: Hello, Alice!
// @expect: Hello, Bob!
function curry3<A, B, C, R>(fn: (a: A, b: B, c: C) => R): (a: A) => (b: B) => (c: C) => R {
    return (a: A) => (b: B) => (c: C) => fn(a, b, c);
}

function volume(l: number, w: number, h: number): number {
    return l * w * h;
}

const curriedVolume = curry3(volume);
const boxWithLength10 = curriedVolume(10);
const boxWithLength10Width5 = boxWithLength10(5);

console.log(boxWithLength10Width5(2));
console.log(boxWithLength10Width5(4));
console.log(curriedVolume(3)(4)(5));

function partialApply<T, U, V>(fn: (x: T, y: U) => V, arg1: T): (arg2: U) => V {
    return (arg2: U) => fn(arg1, arg2);
}

function greet(greeting: string, name: string): string {
    return `${greeting}, ${name}!`;
}

const sayHello = partialApply(greet, "Hello");
console.log(sayHello("Alice"));
console.log(sayHello("Bob"));
