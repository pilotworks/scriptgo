class Rectangle {
  private _width: number;
  private _height: number;

  constructor(w: number, h: number) {
    this._width = w;
    this._height = h;
  }

  get width(): number {
    return this._width;
  }

  set width(val: number) {
    if (val > 0) {
      this._width = val;
    }
  }

  get height(): number {
    return this._height;
  }

  set height(val: number) {
    if (val > 0) {
      this._height = val;
    }
  }

  get area(): number {
    return this._width * this._height;
  }

  get perimeter(): number {
    return 2 * (this._width + this._height);
  }

  get isSquare(): boolean {
    return this._width === this._height;
  }
}

const rect = new Rectangle(10, 20);
console.log(rect.width);
console.log(rect.height);
console.log(rect.area);
console.log(rect.perimeter);
console.log(rect.isSquare);

rect.width = 20;
console.log(rect.width);
console.log(rect.area);
console.log(rect.isSquare);

// Invalid negative update ignored
rect.height = -5;
console.log(rect.height);
console.log(rect.area);
