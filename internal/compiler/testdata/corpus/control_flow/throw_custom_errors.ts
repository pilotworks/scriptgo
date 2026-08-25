// @expect: 123
// @expect: TypeError: Not a valid number
// @expect: RangeError: Number must be positive
function parsePositiveInt(s: string): number {
    const val = parseInt(s, 10);
    if (isNaN(val)) {
        throw new TypeError("Not a valid number");
    }
    if (val < 0) {
        throw new RangeError("Number must be positive");
    }
    return val;
}

try {
    console.log(parsePositiveInt("123"));
} catch (e: any) {
    console.log(e.name + ": " + e.message);
}

try {
    console.log(parsePositiveInt("abc"));
} catch (e: any) {
    console.log(e.name + ": " + e.message);
}

try {
    console.log(parsePositiveInt("-50"));
} catch (e: any) {
    console.log(e.name + ": " + e.message);
}
