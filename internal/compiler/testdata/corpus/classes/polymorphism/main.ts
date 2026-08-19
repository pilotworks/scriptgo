class Shape {
  name: string = "Generic Shape";

  area(): number {
    return 0;
  }

  describe(): string {
    return this.name + " area = " + this.area().toString();
  }
}

class Rectangle extends Shape {
  width: number = 0;
  height: number = 0;

  constructor(w: number, h: number) {
    super();
    this.name = "Rectangle";
    this.width = w;
    this.height = h;
  }

  area(): number {
    return this.width * this.height;
  }
}

class Circle extends Shape {
  radius: number = 0;

  constructor(r: number) {
    super();
    this.name = "Circle";
    this.radius = r;
  }

  area(): number {
    return 3.14 * this.radius * this.radius;
  }
}

const s1: Shape = new Rectangle(10, 5);
const s2: Shape = new Circle(2);

console.log(s1.area());
console.log(s1.describe());

console.log(s2.area());
console.log(s2.describe());
