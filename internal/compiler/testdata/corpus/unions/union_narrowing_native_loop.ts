// Flow-sensitive type narrowing and unboxing into native registers
let a: number | undefined = 10;
if (a !== undefined) {
    for (let i = 0; i < 10; i++) {
        a += i;
    }
    // @expect: 55
    console.log(a);
}

let b: string | null = "hello";
if (b !== null) {
    let res = b + " world";
    // @expect: hello world
    console.log(res);
}

let c: boolean | undefined = true;
if (c !== undefined) {
    let flipped = !c;
    // @expect: false
    console.log(flipped);
}

let d: bigint | null = 100n;
if (d !== null) {
    let sum = d + 50n;
    // @expect: 150n
    console.log(sum);
}
