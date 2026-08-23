// @native.expected: cos(0): 1
// @native.expected: sqrt(16): 4

declare function cos(x: number): number;
declare function sqrt(x: number): number;

console.log("cos(0):", cos(0));
console.log("sqrt(16):", sqrt(16));
