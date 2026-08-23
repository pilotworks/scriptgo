// @expect: 1
// @expect: 2
// @expect: 10
const enum Direction {
    Up = 1,
    Down = 2,
    Custom = 10,
}

console.log(Direction.Up);
console.log(Direction.Down);
console.log(Direction.Custom);
