// @expect: 1
// @expect: 2
// @expect: 3
// @expect: 4
function* subGen() {
    yield 2;
    yield 3;
}

function* mainGen() {
    yield 1;
    yield* subGen();
    yield 4;
}

for (const val of mainGen()) {
    console.log(val);
}
