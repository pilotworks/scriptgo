// @expect: Error: bad argument
// @expect: Date: 2023
// @expect: RegExp: ^[a-z]+$
function formatObject(obj: Error | Date | RegExp): string {
    if (obj instanceof Error) {
        return "Error: " + obj.message;
    } else if (obj instanceof Date) {
        return "Date: " + obj.getFullYear();
    } else if (obj instanceof RegExp) {
        return "RegExp: " + obj.source;
    }
    return "Unknown";
}

console.log(formatObject(new Error("bad argument")));
console.log(formatObject(new Date(1700000000000)));
console.log(formatObject(/^[a-z]+$/));
