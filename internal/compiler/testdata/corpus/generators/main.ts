function* numbers() {
    yield 10;
    yield 20;
    yield 30;
}

const g = numbers();
const r1 = g.next();
console.log(r1.value);
console.log(r1.done);

const r2 = g.next();
console.log(r2.value);
console.log(r2.done);

const r3 = g.next();
console.log(r3.value);
console.log(r3.done);

const r4 = g.next();
console.log(r4.done);

function* countTo(a: number, b: number) {
    yield a;
    yield b;
}

for (const n of countTo(100, 200)) {
    console.log(n);
}
