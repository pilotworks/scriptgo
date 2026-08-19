interface Shape {
  area(): number;
}

abstract class BaseShape implements Shape {
  name: string = "";

  constructor(name: string) {
    this.name = name;
  }

  abstract area(): number;

  describe(): string {
    return "Shape " + this.name + " has area: " + this.area().toString();
  }
}

class Square extends BaseShape {
  side: number = 0;

  constructor(name: string, side: number) {
    super(name);
    this.side = side;
  }

  area(): number {
    return this.side * this.side;
  }
}

const sq = new Square("Box", 4);
console.log(sq.describe());
