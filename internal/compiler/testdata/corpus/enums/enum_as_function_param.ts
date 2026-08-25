// @expect: Moving UP by 5 steps
// @expect: Moving RIGHT by 12 steps
enum Direction {
    Up = "UP",
    Down = "DOWN",
    Left = "LEFT",
    Right = "RIGHT"
}

function move(dir: Direction, steps: number): string {
    return "Moving " + dir + " by " + steps + " steps";
}

console.log(move(Direction.Up, 5));
console.log(move(Direction.Right, 12));
