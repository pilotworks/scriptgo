class Circle {
    kind: string = "circle";
    radius: number;
    constructor(r: number) {
        this.radius = r;
    }
}

class Square {
    kind: string = "square";
    size: number;
    constructor(s: number) {
        this.size = s;
    }
}

type Shape = Circle | Square;

const c: Shape = new Circle(10);
if (c.kind === "circle") {
    console.log(c.radius);
}

const sq: Shape = new Square(5);
if (sq.kind === "square") {
    console.log(sq.size);
}
