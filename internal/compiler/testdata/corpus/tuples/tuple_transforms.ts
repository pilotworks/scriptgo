// @expect: 20
// @expect: 10
// @expect: Alice: 95
// @expect: Bob: 80
type Point2D = [number, number];

function swap(p: Point2D): Point2D {
    return [p[1], p[0]];
}

const original: Point2D = [10, 20];
const inverted = swap(original);
console.log(inverted[0]);
console.log(inverted[1]);

type StudentGrade = [string, number];
const students: StudentGrade[] = [
    ["Alice", 95],
    ["Bob", 80]
];

for (const s of students) {
    console.log(s[0] + ": " + s[1]);
}
