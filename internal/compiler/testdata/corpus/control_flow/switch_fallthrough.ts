// @expect: one two three
// @expect: two three
// @expect: three
// @expect: other
function testFallthrough(val: number): string {
    let result = "";
    switch (val) {
        case 1:
            result += "one ";
        case 2:
            result += "two ";
        case 3:
            result += "three";
            break;
        default:
            result = "other";
    }
    return result;
}

console.log(testFallthrough(1));
console.log(testFallthrough(2));
console.log(testFallthrough(3));
console.log(testFallthrough(4));
