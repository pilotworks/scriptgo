class Rectangle {
  w: number = 0;
  h: number = 0;

  constructor(w: number, h: number) {
    this.w = w;
    this.h = h;
  }

  get area(): number {
    return this.w * this.h;
  }

  set width(val: number) {
    this.w = val;
  }
}

const rect = new Rectangle(5, 10);
console.log(rect.area);
rect.width = 8;
console.log(rect.area);
