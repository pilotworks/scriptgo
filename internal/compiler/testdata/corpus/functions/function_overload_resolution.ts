// @expect: Number: 42.50
// @expect: String: test
function formatValue(x: number): string;
function formatValue(x: string): string;
function formatValue(x: unknown): string {
    if (typeof x === "number") {
        return "Number: " + x.toFixed(2);
    }
    return "String: " + x;
}

console.log(formatValue(42.5));
console.log(formatValue("test"));
