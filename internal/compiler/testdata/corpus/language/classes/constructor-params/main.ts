class Point {
  x: number = 0;
  y: number = 0;

  constructor(x: number, y: number) {
    this.x = x;
    this.y = y;
  }

  sum(): number {
    return this.x + this.y;
  }
}

const p: Point = new Point(15, 27);
console.log(p.sum());
