class Point {
    x: number;
    y: number;
    constructor(x: number, y: number) {
        this.x = x;
        this.y = y;
    }
}

const pt: Point = new Point(10, 25);
const { x, y } = pt;
console.log(x);
console.log(y);

const numbers: number[] = [100, 200, 300, 400];
const [first, second, ...rest] = numbers;
console.log(first);
console.log(second);
console.log(rest.length);
console.log(rest[0]);
console.log(rest[1]);
