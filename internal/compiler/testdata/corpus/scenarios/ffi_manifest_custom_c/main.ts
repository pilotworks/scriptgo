// @native.expected: custom_add(15, 27): 42
// @native.expected: custom_scale(3.5, 4): 14

declare function custom_add(a: number, b: number): number;
declare function custom_scale(val: number, factor: number): number;

console.log("custom_add(15, 27):", custom_add(15, 27));
console.log("custom_scale(3.5, 4):", custom_scale(3.5, 4));
