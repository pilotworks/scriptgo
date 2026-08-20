const e1 = new Error("something went wrong");
console.log(e1.name);
console.log(e1.message);

const e2 = new TypeError("invalid argument type");
console.log(e2.name);
console.log(e2.message);

const e3 = new RangeError("index out of range");
console.log(e3.name);
console.log(e3.message);

const e4 = new SyntaxError("unexpected token");
console.log(e4.name);
console.log(e4.message);

try {
    throw new TypeError("thrown type error");
} catch (err) {
    console.log("caught error");
}
