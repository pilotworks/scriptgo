// @expect: Rectangle: 20
// @expect: Circle: 28
abstract class Shape {
    abstract area(): number;
    abstract name(): string;
}

class Rectangle extends Shape {
    constructor(public width: number, public height: number) {
        super();
    }

    area(): number {
        return this.width * this.height;
    }

    name(): string {
        return "Rectangle";
    }
}

class Circle extends Shape {
    constructor(public radius: number) {
        super();
    }

    area(): number {
        return Math.floor(3.14159 * this.radius * this.radius);
    }

    name(): string {
        return "Circle";
    }
}

const shapes: Shape[] = [new Rectangle(4, 5), new Circle(3)];

for (const s of shapes) {
    console.log(s.name() + ": " + s.area());
}
