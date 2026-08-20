class Point {
    x: number;
    y: number;
    constructor(x: number, y: number) {
        this.x = x;
        this.y = y;
    }
}

const points: Point[] = [new Point(1, 2), new Point(10, 20)];
for (const { x, y } of points) {
    console.log(x);
    console.log(y);
}

const pairs: [string, number][] = [["apple", 5], ["banana", 10]];
for (const [fruit, count] of pairs) {
    console.log(fruit);
    console.log(count);
}
