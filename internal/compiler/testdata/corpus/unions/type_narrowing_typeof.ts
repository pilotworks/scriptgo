// @expect: str: HELLO
// @expect: num: 84
// @expect: bool: TRUE
// @expect: bool: FALSE
function processValue(val: unknown): string {
    if (typeof val === "string") {
        return "str: " + (val as string).toUpperCase();
    } else if (typeof val === "number") {
        return "num: " + ((val as number) * 2);
    } else {
        return "bool: " + (val ? "TRUE" : "FALSE");
    }
}

console.log(processValue("hello"));
console.log(processValue(42));
console.log(processValue(true));
console.log(processValue(false));
