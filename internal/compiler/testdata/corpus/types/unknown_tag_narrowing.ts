// @expect: number: 84
// @expect: string: HELLO
// @expect: boolean: false
// @expect: null value
// @expect: undefined value
// @expect: array of len: 4
function processUnknown(val: unknown): string {
    if (typeof val === "number") {
        return "number: " + (val * 2);
    }
    if (typeof val === "string") {
        return "string: " + val.toUpperCase();
    }
    if (typeof val === "boolean") {
        return "boolean: " + (!val);
    }
    if (val === null) {
        return "null value";
    }
    if (val === undefined) {
        return "undefined value";
    }
    if (Array.isArray(val)) {
        return "array of len: " + val.length;
    }
    return "other";
}

const items: unknown[] = [42, "hello", true, null, undefined, [1, 2, 3, 4]];

for (let i = 0; i < items.length; i++) {
    console.log(processUnknown(items[i]));
}
