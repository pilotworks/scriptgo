// @expect: Point(0, 0)
// @expect: Point(5, 10)
// @expect: Polar Point(3, 4)
class Point2D {
    readonly x: number;
    readonly y: number;

    constructor(x?: number, y?: number) {
        this.x = x ?? 0;
        this.y = y ?? 0;
    }

    static fromPolar(r: number, theta: number): Point2D {
        const x = Math.round(r * Math.cos(theta));
        const y = Math.round(r * Math.sin(theta));
        return new Point2D(x, y);
    }

    toString(): string {
        return "Point(" + this.x + ", " + this.y + ")";
    }
}

const origin = new Point2D();
console.log(origin.toString());

const p1 = new Point2D(5, 10);
console.log(p1.toString());

const p2 = Point2D.fromPolar(5, 0.927295218);
console.log("Polar " + p2.toString());
