class Point {
    x: number;
    y: number;
    constructor(x: number, y: number) {
        this.x = x;
        this.y = y;
    }
}
const p = new Point(1, 2);
Object.assign(p, new Point(10, 20));
console.log(p.x + "," + p.y);
