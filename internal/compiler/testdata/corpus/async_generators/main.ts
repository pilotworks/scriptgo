function* subGen() {
    yield 100;
    yield 200;
}

function* mainGen() {
    yield 10;
    yield* subGen();
    yield 20;
}

for (const n of mainGen()) {
    console.log(n);
}

async function* asyncNumbers() {
    yield 1;
    yield 2;
}

for await (const x of asyncNumbers()) {
    console.log(x);
}

export {};

