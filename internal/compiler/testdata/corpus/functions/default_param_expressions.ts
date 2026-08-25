// @expect: rect_5x10 (5x10)
// @expect: rect_5x10 (5x10)
// @expect: custom (5x10)
function createRectangle(width: number, height: number = width * 2, label: string = "rect_" + width + "x" + height): string {
    return label + " (" + width + "x" + height + ")";
}

console.log(createRectangle(5));
console.log(createRectangle(5, 10));
console.log(createRectangle(5, 10, "custom"));
