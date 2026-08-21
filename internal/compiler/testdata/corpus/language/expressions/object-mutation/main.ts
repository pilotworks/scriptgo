class Box {
  width: number = 10;
  height: number = 20;
}

const b: Box = new Box();
b.width = 50;
console.log(b.width);
